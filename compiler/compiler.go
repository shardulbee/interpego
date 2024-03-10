package compiler

import (
	"fmt"

	"interpego/ast"
	"interpego/code"
	"interpego/object"
)

type Compiler struct {
	instructions code.Instructions
	constants    []object.Object
}

func New() *Compiler {
	return &Compiler{code.Instructions{}, []object.Object{}}
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
	}

	return nil
}

func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{Instructions: c.instructions, Constants: c.constants}
}

func (c *Compiler) addConstant(obj object.Object) int {
	c.constants = append(c.constants, obj)
	return len(c.constants) - 1
}

func (c *Compiler) emit(op code.Opcode, operands ...int) int {
	return c.addInstruction(code.Make(op, operands...))
}

func (c *Compiler) addInstruction(ins code.Instructions) int {
	c.instructions = append(c.instructions, ins...)
	return len(ins)
}
