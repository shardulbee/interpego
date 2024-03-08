package evaluator

import (
	"fmt"

	"interpego/ast"
	"interpego/object"
)

var (
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
	NULL  = &object.Null{}
)

func Eval(node ast.Node, env *object.Environment) object.Object {
	switch node := node.(type) {
	case *ast.Program:
		return evalProgram(env, node)
	case *ast.BlockStatement:
		return evalBlockStatement(env, node)
	case *ast.ExpressionStatement:
		return Eval(node.Expression, env)
	case *ast.ForLoop:
		return evalForLoop(env, node)
	case *ast.LetStatement:
		value := Eval(node.Value, env)
		if isError(value) {
			return value
		}
		env.Set(node.Name.Value, value)
		return value
	case *ast.ReturnStatement:
		obj := Eval(node.ReturnValue, env)
		if isError(obj) {
			return obj
		}
		return &object.ReturnValue{Value: obj}
	case *ast.PrefixExpression:
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalPrefixExpression(node.Operator, right)
	case *ast.InfixExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalInfixExpression(left, node.Operator, right)
	case *ast.IfExpression:
		condition := Eval(node.Condition, env)
		if isError(condition) {
			return condition
		}
		return evalIfElseExpression(env, condition, node.Consequence, node.Alternative)
	case *ast.FunctionLiteral:
		return &object.Function{
			Env:    env,
			Params: node.Parameters,
			Body:   node.FunctionBody,
		}
	case *ast.CallExpression:
		evaluatedArgs := evaluateCallArguments(env, node.Arguments)
		if len(evaluatedArgs) == 1 && isError(evaluatedArgs[0]) {
			return evaluatedArgs[0]
		}

		result := Eval(node.Function, env)
		if isError(result) {
			return result
		}
		switch fn := result.(type) {
		case *object.Function:
			return evalCallExpression(fn, evaluatedArgs)
		case *object.Builtin:
			return fn.Fn(evaluatedArgs...)
		default:
			return newError("not a function: %s", fn.Type())
		}
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}
	case *ast.StringLiteral:
		return &object.String{Value: node.Value}
	case *ast.BooleanLiteral:
		return nativeBoolToBooleanObject(node.Value)
	case *ast.ArrayLiteral:
		elements := evaluateCallArguments(env, node.Elements)
		if len(elements) == 1 && isError(elements[0]) {
			return elements[0]
		}

		return &object.Array{Elements: elements}
	case *ast.HashLiteral:
		pairs := node.Pairs
		evaluatedPairs := make(map[object.HashKey]object.HashPair)
		for keyNode := range pairs {
			evaluatedKey := Eval(keyNode, env)
			if isError(evaluatedKey) {
				return evaluatedKey
			}
			hashable, ok := evaluatedKey.(object.Hashable)
			if !ok {
				return newError("key type is not hashable: %s", evaluatedKey.Type())
			}
			evaluatedValue := Eval(pairs[keyNode], env)
			if isError(evaluatedKey) {
				return evaluatedKey
			}
			if isError(evaluatedValue) {
				return evaluatedValue
			}
			evaluatedPairs[hashable.HashKey()] = object.HashPair{Key: evaluatedKey, Value: evaluatedValue}
		}
		return &object.Hash{Pairs: evaluatedPairs}
	case *ast.IndexExpression:
		idx := Eval(node.Index, env)
		if isError(idx) {
			return idx
		}

		arr := Eval(node.Left, env)
		if isError(arr) {
			return arr
		}

		return evalIndexExpression(arr, idx)
	case *ast.Identifier:
		if val, ok := env.Get(node.Value); ok {
			return val
		}
		if builtin, ok := NewEvaluator().Builtins[node.Value]; ok {
			return builtin
		}

		return newError("unknown identifier: %s", node.Value)
	}
	return newError("default branch of eval. could not handle: %T", node)
}

func evalInfixExpression(left object.Object, operator string, right object.Object) object.Object {
	switch {
	case left.Type() == object.INTEGER_TYPE && right.Type() == object.INTEGER_TYPE:
		return evalIntegerInfixExpression(left.(*object.Integer), operator, right.(*object.Integer))
	case left.Type() == object.STRING_TYPE && right.Type() == object.STRING_TYPE:
		return evalStringInfixExpression(left.(*object.String), operator, right.(*object.String))
	case left.Type() == object.ARRAY_TYPE && right.Type() == object.ARRAY_TYPE:
		return evalArrayInfixExpression(left.(*object.Array), operator, right.(*object.Array))
	case operator == "!=":
		return nativeBoolToBooleanObject(left != right)
	case operator == "==":
		return nativeBoolToBooleanObject(left == right)
	case left.Type() != right.Type():
		return newError("type mismatch: %s %s %s", left.Type(), operator, right.Type())
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
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
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalStringInfixExpression(left *object.String, operator string, right *object.String) object.Object {
	switch operator {
	case "+":
		return &object.String{Value: left.Value + right.Value}
	case "<":
		return nativeBoolToBooleanObject(left.Value < right.Value)
	case ">":
		return nativeBoolToBooleanObject(left.Value > right.Value)
	case "==":
		return nativeBoolToBooleanObject(left.Value == right.Value)
	case "!=":
		return nativeBoolToBooleanObject(left.Value != right.Value)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalArrayInfixExpression(left *object.Array, operator string, right *object.Array) object.Object {
	switch operator {
	case "+":
		concatenated := make([]object.Object, len(left.Elements)+len(right.Elements))
		copy(concatenated, left.Elements)
		copy(concatenated[len(left.Elements):], right.Elements)
		return &object.Array{Elements: concatenated}
	case "==":
		if len(left.Elements) != len(right.Elements) {
			return FALSE
		}
		for i, elem := range left.Elements {
			res := evalInfixExpression(elem, "==", right.Elements[i])
			switch res {
			case TRUE:
				continue
			case FALSE:
				return FALSE
			default:
				if isError(res) {
					return res
				}
				return FALSE
			}
		}
		return TRUE
	case "!=":
		if len(left.Elements) != len(right.Elements) {
			return TRUE
		}
		for i, elem := range left.Elements {
			res := evalInfixExpression(elem, "!=", right.Elements[i])
			switch res {
			case TRUE:
				return TRUE
			case FALSE:
				continue
			default:
				if isError(res) {
					return res
				}
				return newError("unexpected value when comparing elements: got=%T (%+v)", res, res)
			}
		}
		return FALSE
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalPrefixExpression(operator string, right object.Object) object.Object {
	switch operator {
	case "!":
		return evalBangOperatorExpression(right)
	case "-":
		return evalMinusOperatorExpression(right)
	default:
		return newError("unknown operator: %s%s", operator, right.Type())
	}
}

func evalMinusOperatorExpression(exp object.Object) object.Object {
	if exp.Type() != object.INTEGER_TYPE {
		return newError("unknown operator: -%s", exp.Type())
	}
	return &object.Integer{Value: -exp.(*object.Integer).Value}
}

func evalBangOperatorExpression(right object.Object) object.Object {
	boolean, ok := right.(*object.Boolean)
	if !ok {
		return newError("unknown operator: !%s", right.Type())
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

func evalProgram(env *object.Environment, program *ast.Program) object.Object {
	var result object.Object
	for _, stmt := range program.Statements {
		result = Eval(stmt, env)
		switch result := result.(type) {
		case *object.ReturnValue:
			return result.Value
		case *object.Error:
			return result
		}
	}
	return result
}

func evalBlockStatement(env *object.Environment, bs *ast.BlockStatement) object.Object {
	var result object.Object
	for _, stmt := range bs.Statements {
		result = Eval(stmt, env)
		if result != nil {
			rt := result.Type()
			if rt == object.RETURN_TYPE || rt == object.ERROR_TYPE {
				return result
			}
		}
	}
	return result
}

func evalIfElseExpression(env *object.Environment, condition object.Object, consequence *ast.BlockStatement, alternative *ast.BlockStatement) object.Object {
	if isTruthy(condition) {
		return Eval(consequence, env)
	} else if alternative == nil {
		return NULL
	}
	return Eval(alternative, env)
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

func newError(format string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}

func isError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.ERROR_TYPE
	}
	return false
}

func evaluateCallArguments(env *object.Environment, args []ast.Expression) []object.Object {
	var evaluatedArgs []object.Object
	for _, arg := range args {
		result := Eval(arg, env)
		if isError(result) {
			return []object.Object{result}
		}
		evaluatedArgs = append(evaluatedArgs, result)
	}
	return evaluatedArgs
}

func evalCallExpression(function *object.Function, args []object.Object) object.Object {
	if len(function.Params) != len(args) {
		return newError(
			"incorrect number of arguments passed to function: expected=%d, got=%d",
			len(function.Params),
			len(args),
		)
	}
	extended := extendFunctionEnvironment(function, args)
	applied := Eval(function.Body, extended)
	return unwrapReturnValue(applied)
}

func unwrapReturnValue(obj object.Object) object.Object {
	if returnVal, ok := obj.(*object.ReturnValue); ok {
		return returnVal.Value
	}
	return obj
}

func extendFunctionEnvironment(function *object.Function, args []object.Object) *object.Environment {
	environment := object.NewEnclosedEnvironment(function.Env)
	for i, param := range function.Params {
		environment.Set(param.Value, args[i])
	}
	return environment
}

func evalIndexExpression(indexable object.Object, idxObj object.Object) object.Object {
	switch {
	case indexable.Type() == object.ARRAY_TYPE && idxObj.Type() == object.INTEGER_TYPE:
		arr := indexable.(*object.Array).Elements
		idx := idxObj.(*object.Integer).Value

		if idx < 0 || idx > int64(len(arr)-1) {
			return newError("array index out of bounds: size=%d, index=%d", len(arr), idx)
		}
		return arr[idx]
	case indexable.Type() == object.HASH_TYPE:
		hash := indexable.(*object.Hash).Pairs
		hashKey, ok := idxObj.(object.Hashable)
		if !ok {
			return newError("unusable as hash key: %s", idxObj.Type())
		}
		if pair, ok := hash[hashKey.HashKey()]; ok {
			return pair.Value
		}
		return NULL
	default:
		return newError("index operator not supported: %s", indexable.Type())
	}
}

func evalForLoop(env *object.Environment, forLoop *ast.ForLoop) object.Object {
	if initResult := Eval(forLoop.InitStatement, env); isError(initResult) {
		return initResult
	}
	evalCondition := Eval(forLoop.Condition, env)
	if isError(evalCondition) {
		return evalCondition
	}
	if evalCondition.Type() != object.BOOLEAN_TYPE {
		return newError("for loop condition must be have type %s", object.BOOLEAN_TYPE)
	}

	conditionResult := evalCondition.(*object.Boolean)
	var forResult object.Object
	for conditionResult.Value {
		forResult = Eval(forLoop.ForBody, env)
		if isError(forResult) {
			return forResult
		}

		if postResult := Eval(forLoop.PostStatement, env); isError(postResult) {
			return postResult
		}
		conditionResult = Eval(forLoop.Condition, env).(*object.Boolean)
	}

	return forResult
}
