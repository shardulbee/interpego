package evaluator

// var builtins = map[string]*object.Builtin{
// 	"first": {
// 		Fn: func(args ...object.Object) object.Object {
// 			if len(args) != 1 {
// 				return newError("wrong number of arguments. got=%d, want=1", len(args))
// 			}
// 			if args[0].Type() != object.ARRAY_TYPE {
// 				return newError("argument to `first` not supported, got=%s, expected=%s", args[0].Type(), object.ARRAY_TYPE)
// 			}

// 			arr := args[0].(*object.Array).Elements
// 			if len(arr) == 0 {
// 				return newError("array index out of bounds: size=%d, index=%d", len(arr), 0)
// 			}
// 			return arr[0]
// 		},
// 	},
// 	"last": {
// 		Fn: func(args ...object.Object) object.Object {
// 			if len(args) != 1 {
// 				return newError("wrong number of arguments. got=%d, want=1", len(args))
// 			}
// 			if args[0].Type() != object.ARRAY_TYPE {
// 				return newError("argument to `first` not supported, got=%s, expected=%s", args[0].Type(), object.ARRAY_TYPE)
// 			}

// 			arr := args[0].(*object.Array).Elements
// 			if len(arr) == 0 {
// 				return newError("array index out of bounds: size=%d, index=%d", len(arr), 0)
// 			}
// 			return arr[len(arr)-1]
// 		},
// 	},
// 	"rest": {
// 		Fn: func(args ...object.Object) object.Object {
// 			if args[0].Type() != object.ARRAY_TYPE {
// 				return newError("argument to `first` not supported, got=%s, expected=%s", args[0].Type(), object.ARRAY_TYPE)
// 			}

// 			arr := args[0].(*object.Array).Elements

// 			if len(arr) > 0 {
// 				newArr := make([]object.Object, len(arr)-1, len(arr)-1)
// 				copy(newArr, arr[1:])
// 				return &object.Array{Elements: newArr}
// 			}

// 			return NULL
// 		},
// 	},
// 	"push": {
// 		Fn: func(args ...object.Object) object.Object {
// 			if len(args) != 2 {
// 				return newError("wrong number of arguments. got=%d, want=2",
// 					len(args))
// 			}

// 			if args[0].Type() != object.ARRAY_TYPE {
// 				return newError("argument to `push` must be ARRAY, got %s", args[0].Type())
// 			}

// 			arr := args[0].(*object.Array).Elements
// 			length := len(arr)
// 			newArr := make([]object.Object, length+1, length+1)
// 			copy(newArr, arr)
// 			newArr[length] = args[1]
// 			return &object.Array{Elements: newArr}
// 		},
// 	},
// 	"print": {
// 		Fn: func(args ...object.Object) object.Object {
// 			if len(args) != 1 {
// 				return newError("wrong number of arguments. got=%d, want=1", len(args))
// 			}
// 			fmt.Printf("%s\n", args[0].Inspect())
// 			return NULL
// 		},
// 	},
// 	"len": {
// 		Fn: func(args ...object.Object) object.Object {
// 			if len(args) != 1 {
// 				return newError("wrong number of arguments. got=%d, want=1", len(args))
// 			}
// 			switch arg := args[0].(type) {
// 			case *object.String:
// 				return &object.Integer{Value: int64(len(arg.Value))}
// 			case *object.Array:
// 				return &object.Integer{Value: int64(len(arg.Elements))}
// 			default:
// 				return newError("argument to `len` not supported, got %s", args[0].Type())
// 			}
// 		},
// 	},
// }
