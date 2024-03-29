package compiler

import (
	"fmt"

	"interpego/ast"
	"interpego/code"
	"interpego/object"
)

type CompilationScope struct {
	instructions    code.Instructions
	lastInstruction EmittedInstruction
	prevInstruction EmittedInstruction
}

type EmittedInstruction struct {
	Opcode   code.Opcode
	Position int
}

type Compiler struct {
	scopes      []CompilationScope
	scopeIdx    int
	constants   []object.Object
	symbolTable *SymbolTable
}

func New() *Compiler {
	mainScope := CompilationScope{
		instructions:    code.Instructions{},
		lastInstruction: EmittedInstruction{},
		prevInstruction: EmittedInstruction{},
	}
	return &Compiler{scopes: []CompilationScope{mainScope}, scopeIdx: 0, constants: []object.Object{}, symbolTable: NewSymbolTable()}
}

func NewWithSymbols(symbols *SymbolTable) *Compiler {
	mainScope := CompilationScope{
		instructions:    code.Instructions{},
		lastInstruction: EmittedInstruction{},
		prevInstruction: EmittedInstruction{},
	}
	return &Compiler{scopes: []CompilationScope{mainScope}, scopeIdx: 0, constants: []object.Object{}, symbolTable: symbols}
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
		sym := c.symbolTable.Define(node.Name.Value)
		if sym.Scope == GLOBAL_SCOPE {
			c.emit(code.OpSetGlobal, sym.Index)
		} else {
			c.emit(code.OpSetLocal, sym.Index)
		}
	case *ast.Identifier:
		sym, ok := c.symbolTable.Resolve(node.Value)
		if !ok {
			return fmt.Errorf("unable to resolve identifier: ident=%s", node.Value)
		}
		if sym.Scope == GLOBAL_SCOPE {
			c.emit(code.OpGetGlobal, sym.Index)
		} else {
			c.emit(code.OpGetLocal, sym.Index)
		}
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
		if c.lastInstructionIs(code.OpPop) {
			c.removeLastPop()
		}

		jumpAlwaysIns := c.emit(code.OpJump, 9999)
		c.changeOperand(jumpNotTruthyIns, len(c.currentInstructions()))

		if node.Alternative != nil {
			c.Compile(node.Alternative)
			if c.lastInstructionIs(code.OpPop) {
				c.removeLastPop()
			}
		} else {
			c.emit(code.OpNull)
		}
		c.changeOperand(jumpAlwaysIns, len(c.currentInstructions()))
	case *ast.CallExpression:
		err := c.Compile(node.Function)
		if err != nil {
			return err
		}
		for _, arg := range node.Arguments {
			err := c.Compile(arg)
			if err != nil {
				return err
			}
		}
		c.emit(code.OpCall, len(node.Arguments))
	case *ast.IntegerLiteral:
		c.emit(code.OpConstant, c.addConstant(&object.Integer{Value: node.Value}))
	case *ast.StringLiteral:
		c.emit(code.OpConstant, c.addConstant(&object.String{Value: node.Value}))
	case *ast.FunctionLiteral:
		c.enterScope()

		for _, param := range node.Parameters {
			c.symbolTable.Define(param.Value)
		}

		err := c.Compile(node.FunctionBody)
		if err != nil {
			return err
		}
		if c.lastInstructionIs(code.OpPop) {
			newIns := code.Make(code.OpReturnValue)
			c.replaceInstruction(c.scopes[c.scopeIdx].lastInstruction.Position, newIns)
			c.scopes[c.scopeIdx].lastInstruction.Opcode = code.OpReturnValue
		}
		if !c.lastInstructionIs(code.OpReturnValue) {
			c.emit(code.OpReturn)
		}
		numLocals := c.symbolTable.numDefinitions
		newIns := c.leaveScope()
		c.emit(
			code.OpConstant,
			c.addConstant(&object.CompiledFunction{
				Instructions: newIns,
				NumLocals:    numLocals,
			}),
		)
	case *ast.ReturnStatement:
		err := c.Compile(node.ReturnValue)
		if err != nil {
			return err
		}
		c.emit(code.OpReturnValue)
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
	default:
		return fmt.Errorf("unsupported ast node type. got=%T (%+v)", node, node)
	}

	return nil
}

func (c *Compiler) lastInstructionIs(op code.Opcode) bool {
	return c.scopes[c.scopeIdx].lastInstruction.Opcode == op
}

// If the last instruction in a block expression is a pop we remove it.
// This allows blocks to implicitly return the evaluated value of the last ExpressionStatement
// in block
func (c *Compiler) maybeRemoveLastPop() bool {
	if c.lastInstructionIs(code.OpPop) {
		c.scopes[c.scopeIdx].instructions = c.currentInstructions()[:c.scopes[c.scopeIdx].lastInstruction.Position]
		c.scopes[c.scopeIdx].lastInstruction = c.scopes[c.scopeIdx].prevInstruction
		return true
	}
	return false
}

func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{Instructions: c.currentInstructions(), Constants: c.constants}
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

	c.scopes[c.scopeIdx].prevInstruction = c.scopes[c.scopeIdx].lastInstruction
	c.scopes[c.scopeIdx].lastInstruction = EmittedInstruction{op, pos}

	return pos
}

func (c *Compiler) currentInstructions() code.Instructions {
	return c.scopes[c.scopeIdx].instructions
}

func (c *Compiler) addInstruction(ins code.Instructions) int {
	posNewInstruction := len(c.currentInstructions())
	updatedInstructions := append(c.currentInstructions(), ins...)
	c.scopes[c.scopeIdx].instructions = updatedInstructions

	return posNewInstruction
}

func (c *Compiler) replaceInstruction(opPos int, newIns []byte) {
	for i := 0; i < len(newIns); i++ {
		c.currentInstructions()[opPos+i] = newIns[i]
	}
}

// changeOperand updates the operand of an existing opcode at a given position.
// opPos is the starting index of the opcode in the instructions slice.
// operand is the new operand value to be used.
func (c *Compiler) changeOperand(opPos int, newOperand int) {
	op := code.Opcode(c.currentInstructions()[opPos])
	newIns := code.Make(op, newOperand)
	c.replaceInstruction(opPos, newIns)
}

func (c *Compiler) enterScope() {
	newScope := CompilationScope{
		instructions:    code.Instructions{},
		lastInstruction: EmittedInstruction{},
		prevInstruction: EmittedInstruction{},
	}
	c.symbolTable = NewNestedSymbolTable(c.symbolTable)
	c.scopes = append(c.scopes, newScope)
	c.scopeIdx++
}

func (c *Compiler) leaveScope() code.Instructions {
	if c.scopeIdx == 0 {
		panic("attempting to leave global scope")
	}
	ins := c.currentInstructions()
	c.scopes = c.scopes[:c.scopeIdx]
	c.scopeIdx--
	c.symbolTable = c.symbolTable.outer
	return ins
}

func (c *Compiler) removeLastPop() {
	c.scopes[c.scopeIdx].instructions = c.currentInstructions()[:c.scopes[c.scopeIdx].lastInstruction.Position]
	c.scopes[c.scopeIdx].lastInstruction = c.scopes[c.scopeIdx].prevInstruction
}
