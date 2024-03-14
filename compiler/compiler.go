package compiler

import (
	"fmt"

	"interpego/ast"
	"interpego/code"
	"interpego/object"
)

type EmittedInstruction struct {
	Opcode   code.Opcode
	Position int
}

type Compiler struct {
	instructions    code.Instructions
	constants       []object.Object
	symbolTable     *SymbolTable
	lastInstruction EmittedInstruction
	prevInstruction EmittedInstruction
}

func New() *Compiler {
	return &Compiler{code.Instructions{}, []object.Object{}, NewSymbolTable(), EmittedInstruction{}, EmittedInstruction{}}
}

type Bytecode struct {
	Instructions code.Instructions
	Constants    []object.Object
}

func (c *Compiler) Compile(node ast.Node) error {
	switch node := node.(type) {
	case *ast.ExpressionStatement:
		err := c.Compile(node.Expression)
		if err != nil {
			return err
		}
		c.emit(code.OpPop)
	case *ast.LetStatement:
		err := c.Compile(node.Value)
		if err != nil {
			return err
		}
		sym := c.symbolTable.Set(node.Name.Value)
		c.emit(code.OpSetGlobal, sym.Index)
	case *ast.Identifier:
		sym, ok := c.symbolTable.Get(node.Value)
		if !ok {
			return fmt.Errorf("unable to resolve identifier: ident=%s", node.Value)
		}
		c.emit(code.OpGetGlobal, sym.Index)
	case *ast.InfixExpression:
		if node.Operator == "<" {
			err := c.Compile(node.Right)
			if err != nil {
				return err
			}

			err = c.Compile(node.Left)
			if err != nil {
				return err
			}
			c.emit(code.OpGreaterThan)
			return nil
		}
		err := c.Compile(node.Left)
		if err != nil {
			return err
		}

		err = c.Compile(node.Right)
		if err != nil {
			return err
		}

		switch node.Operator {
		case "+":
			c.emit(code.OpAdd)
		case "-":
			c.emit(code.OpSub)
		case "*":
			c.emit(code.OpMul)
		case "/":
			c.emit(code.OpDiv)
		case "==":
			c.emit(code.OpEqual)
		case "!=":
			c.emit(code.OpNotEqual)
		case ">":
			c.emit(code.OpGreaterThan)
		case "<":
			c.emit(code.OpGreaterThan)

		default:
			return fmt.Errorf("unsupported operator %q", node.Operator)
		}
	case *ast.PrefixExpression:
		switch node.Operator {
		case "!":
			err := c.Compile(node.Right)
			if err != nil {
				return err
			}
			c.emit(code.OpBang)
		case "-":
			err := c.Compile(node.Right)
			if err != nil {
				return err
			}
			c.emit(code.OpMinus)
		default:
			return fmt.Errorf("unknown operator: %q", node.Operator)
		}
		return nil
	case *ast.IfExpression:
		err := c.Compile(node.Condition)
		if err != nil {
			return err
		}

		jumpNotTruthyIns := c.emit(code.OpJumpNotTruthy, 9999)
		c.Compile(node.Consequence)
		jumpAlwaysIns := c.emit(code.OpJump, 9999)
		c.changeOperand(jumpNotTruthyIns, len(c.instructions))

		if node.Alternative != nil {
			c.Compile(node.Alternative)
		} else {
			c.emit(code.OpNull)
		}
		c.changeOperand(jumpAlwaysIns, len(c.instructions))
	case *ast.IntegerLiteral:
		c.emit(code.OpConstant, c.addConstant(&object.Integer{Value: node.Value}))
	case *ast.StringLiteral:
		c.emit(code.OpConstant, c.addConstant(&object.String{Value: node.Value}))
	case *ast.BooleanLiteral:
		if node.Value {
			c.emit(code.OpTrue)
		} else {
			c.emit(code.OpFalse)
		}
	case *ast.Program:
		stmts := node.Statements
		for _, stmt := range stmts {
			err := c.Compile(stmt)
			if err != nil {
				return err
			}
		}
	case *ast.BlockStatement:
		stmts := node.Statements
		for _, stmt := range stmts {
			err := c.Compile(stmt)
			if err != nil {
				return err
			}
		}
		c.maybeRemoveLastPop()
	default:
		return fmt.Errorf("unsupported ast node type. got=%T (%+v)", node, node)
	}

	return nil
}

// If the last instruction in a block expression is a pop we remove it.
// This allows blocks to implicitly return the evaluated value of the last ExpressionStatement
// in block
func (c *Compiler) maybeRemoveLastPop() bool {
	if c.lastInstruction.Opcode == code.OpPop {
		c.instructions = c.instructions[:c.lastInstruction.Position]
		c.lastInstruction = c.prevInstruction
		return true
	}
	return false
}

func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{Instructions: c.instructions, Constants: c.constants}
}

func (c *Compiler) addConstant(obj object.Object) int {
	c.constants = append(c.constants, obj)
	return len(c.constants) - 1
}

// emit creates an instruction using the provided opcode and operands.
// It then adds this instruction to the compiler's instruction slice.
// Returns the starting position of the newly added instruction.
func (c *Compiler) emit(op code.Opcode, operands ...int) int {
	newInstruction := code.Make(op, operands...)
	pos := c.addInstruction(newInstruction)

	c.prevInstruction = c.lastInstruction
	c.lastInstruction = EmittedInstruction{op, pos}
	return pos
}

func (c *Compiler) addInstruction(ins code.Instructions) int {
	posNewInstruction := len(c.instructions)
	c.instructions = append(c.instructions, ins...)

	return posNewInstruction
}

func (c *Compiler) replaceInstruction(opPos int, newIns []byte) {
	for i := 0; i < len(newIns); i++ {
		c.instructions[opPos+i] = newIns[i]
	}
}

// changeOperand updates the operand of an existing opcode at a given position.
// opPos is the starting index of the opcode in the instructions slice.
// operand is the new operand value to be used.
func (c *Compiler) changeOperand(opPos int, newOperand int) {
	op := code.Opcode(c.instructions[opPos])
	fmt.Printf("new opCode: %d\n", op)
	newIns := code.Make(op, newOperand)
	c.replaceInstruction(opPos, newIns)
}
