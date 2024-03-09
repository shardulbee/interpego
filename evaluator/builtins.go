package evaluator

import (
	"fmt"

	"interpego/object"
)

type Builtins map[string]*object.Builtin

func NewBuiltins() Builtins {
	return Builtins{
		"first": &object.Builtin{
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return newError("wrong number of arguments. got=%d, want=1", len(args))
				}
				if args[0].Type() != object.ARRAY_TYPE {
					return newError("argument to `first` not supported, got=%s, expected=%s", args[0].Type(), object.ARRAY_TYPE)
				}

				arr := args[0].(*object.Array).Elements
				if len(arr) == 0 {
					return newError("array index out of bounds: size=%d, index=%d", len(arr), 0)
				}
				return arr[0]
			},
		},
		"last": &object.Builtin{
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return newError("wrong number of arguments. got=%d, want=1", len(args))
				}
				if args[0].Type() != object.ARRAY_TYPE {
					return newError("argument to `first` not supported, got=%s, expected=%s", args[0].Type(), object.ARRAY_TYPE)
				}

				arr := args[0].(*object.Array).Elements
				if len(arr) == 0 {
					return newError("array index out of bounds: size=%d, index=%d", len(arr), 0)
				}
				return arr[len(arr)-1]
			},
		},
		"rest": &object.Builtin{
			Fn: func(args ...object.Object) object.Object {
				if args[0].Type() != object.ARRAY_TYPE {
					return newError("argument to `first` not supported, got=%s, expected=%s", args[0].Type(), object.ARRAY_TYPE)
				}

				arr := args[0].(*object.Array).Elements

				if len(arr) > 0 {
					newArr := make([]object.Object, len(arr)-1, len(arr)-1)
					copy(newArr, arr[1:])
					return &object.Array{Elements: newArr}
				}

				return &object.Null{}
			},
		},
		"push": &object.Builtin{
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return newError("wrong number of arguments. got=%d, want=2",
						len(args))
				}

				if args[0].Type() != object.ARRAY_TYPE {
					return newError("argument to `push` must be ARRAY, got %s", args[0].Type())
				}

				arr := args[0].(*object.Array).Elements
				length := len(arr)
				newArr := make([]object.Object, length+1, length+1)
				copy(newArr, arr)
				newArr[length] = args[1]
				return &object.Array{Elements: newArr}
			},
		},
		"print": &object.Builtin{
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return newError("wrong number of arguments. got=%d, want=1", len(args))
				}
				fmt.Printf("%s\n", args[0].Inspect())
				return &object.Null{}
			},
		},
		"len": &object.Builtin{
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 1 {
					return newError("wrong number of arguments. got=%d, want=1", len(args))
				}
				switch arg := args[0].(type) {
				case *object.String:
					return &object.Integer{Value: int64(len(arg.Value))}
				case *object.Array:
					return &object.Integer{Value: int64(len(arg.Elements))}
				default:
					return newError("argument to `len` not supported, got %s", args[0].Type())
				}
			},
		},
		"map": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 2 {
					return newError("wrong number of arguments. got=%d, want=1", len(args))
				}
				fn, ok := args[0].(*object.Function)
				if !ok {
					return newError("expected first argument to `map` to be an object.Function. got=%T", args[0])
				}
				arr, ok := args[1].(*object.Array)
				if !ok {
					return newError("expected second argument to `map` to be an object.Array. got=%T", args[1])
				}
				newArr := make([]object.Object, len(arr.Elements))
				for i, elem := range arr.Elements {
					newElem := evalCallExpression(NewBuiltins(), fn, []object.Object{elem})
					if isError(newElem) {
						return newElem
					}

					newArr[i] = newElem
				}

				return &object.Array{Elements: newArr}
			},
		},
		"reduce": {
			Fn: func(args ...object.Object) object.Object {
				if len(args) != 3 {
					return newError("wrong number of arguments. got=%d, want=1", len(args))
				}
				fn, ok := args[0].(*object.Function)
				if !ok {
					return newError("expected first argument to `map` to be an object.Function. got=%T", args[0])
				}
				arr, ok := args[1].(*object.Array)
				if !ok {
					return newError("expected second argument to `map` to be an object.Array. got=%T", args[1])
				}
				if len(fn.Params) != 2 {
					return newError("function provided to `reduce` must accept two params. got=%d", len(fn.Params))
				}

				acc := args[2]
				for _, elem := range arr.Elements {
					acc = evalCallExpression(NewBuiltins(), fn, []object.Object{elem, acc})
					if isError(acc) {
						return acc
					}
				}
				return acc
			},
		},
	}
}
