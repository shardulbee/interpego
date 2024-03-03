package evaluator

import (
	"interpego/ast"
	"interpego/object"
)

var (
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
	NULL  = &object.Null{}
)

func Eval(node ast.Node) object.Object {
	switch node := node.(type) {
	case *ast.Program:
		return evalProgram(node)
	case *ast.BlockStatement:
		return evalBlockStatement(node)
	case *ast.ExpressionStatement:
		return Eval(node.Expression)
	case *ast.PrefixExpression:
		right := Eval(node.Right)
		return evalPrefixExpression(node.Operator, right)
	case *ast.InfixExpression:
		left := Eval(node.Left)
		right := Eval(node.Right)
		return evalInfixExpression(left, node.Operator, right)
	case *ast.IfExpression:
		condition := Eval(node.Condition)
		return evalIfElseExpression(condition, node.Consequence, node.Alternative)
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}
	case *ast.BooleanLiteral:
		return nativeBoolToBooleanObject(node.Value)
	}

	return NULL
}

func evalInfixExpression(left object.Object, operator string, right object.Object) object.Object {
	switch {
	case left.Type() == object.INTEGER_TYPE && right.Type() == object.INTEGER_TYPE:
		return evalIntegerInfixExpression(left.(*object.Integer), operator, right.(*object.Integer))
	case operator == "!=":
		return nativeBoolToBooleanObject(left != right)
	case operator == "==":
		return nativeBoolToBooleanObject(left == right)
	default:
		return NULL
	}
}

func evalIntegerInfixExpression(left *object.Integer, operator string, right *object.Integer) object.Object {
	switch operator {
	case "+":
		return &object.Integer{Value: left.Value + right.Value}
	case "-":
		return &object.Integer{Value: left.Value - right.Value}
	case "*":
		return &object.Integer{Value: left.Value * right.Value}
	case "/":
		return &object.Integer{Value: left.Value / right.Value}
	case "<":
		return nativeBoolToBooleanObject(left.Value < right.Value)
	case ">":
		return nativeBoolToBooleanObject(left.Value > right.Value)
	case "==":
		return nativeBoolToBooleanObject(left.Value == right.Value)
	case "!=":
		return nativeBoolToBooleanObject(left.Value != right.Value)
	default:
		return NULL
	}
}

func evalPrefixExpression(operator string, right object.Object) object.Object {
	switch operator {
	case "!":
		return evalBangOperatorExpression(right)
	case "-":
		return evalMinusOperatorExpression(right)
	default:
		return NULL
	}
}

func evalMinusOperatorExpression(exp object.Object) object.Object {
	if exp.Type() != object.INTEGER_TYPE {
		return NULL
	}
	return &object.Integer{Value: -exp.(*object.Integer).Value}
}

func evalBangOperatorExpression(right object.Object) object.Object {
	boolean, ok := right.(*object.Boolean)
	if !ok {
		return NULL
	}

	if boolean.Value {
		return FALSE
	}
	return TRUE
}

func nativeBoolToBooleanObject(input bool) *object.Boolean {
	if input {
		return TRUE
	}
	return FALSE
}

func evalProgram(program *ast.Program) object.Object {
	var obj object.Object
	for _, stmt := range program.Statements {
		obj = Eval(stmt)
	}
	return obj
}

func evalBlockStatement(bs *ast.BlockStatement) object.Object {
	var obj object.Object
	for _, stmt := range bs.Statements {
		obj = Eval(stmt)
	}
	return obj
}

func evalIfElseExpression(condition object.Object, consequence *ast.BlockStatement, alternative *ast.BlockStatement) object.Object {
	if isTruthy(condition) {
		return Eval(consequence)
	} else if alternative == nil {
		return NULL
	}
	return Eval(alternative)
}

func isTruthy(condition object.Object) bool {
	switch condition {
	case TRUE:
		return true
	case FALSE:
		return false
	case NULL:
		return false
	default:
		return true
	}
}
