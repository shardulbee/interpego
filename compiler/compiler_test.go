package compiler

import (
	"fmt"
	"testing"

	"interpego/ast"
	"interpego/code"
	"interpego/lexer"
	"interpego/object"
	"interpego/parser"
)

type compilerTestCase struct {
	input                string
	expectedConstants    []interface{}
	expectedInstructions []code.Instructions
}

type expectedCompiledFunction struct {
	instructions  []code.Instructions
	numLocals     int
	numParameters int
}

func TestBooleanExpressions(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "true",
			expectedConstants: []interface{}{},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpTrue),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "false",
			expectedConstants: []interface{}{},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpFalse),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "true < false",
			expectedConstants: []interface{}{},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpFalse),
				code.Make(code.OpTrue),
				code.Make(code.OpGreaterThan),
				code.Make(code.OpPop),
			},
		},
		{
			input: "1 > 2", expectedConstants: []interface{}{1, 2}, expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpGreaterThan),
				code.Make(code.OpPop),
			},
		},
		{
			input: "1 < 2", expectedConstants: []interface{}{2, 1}, expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpGreaterThan),
				code.Make(code.OpPop),
			},
		},
		{
			input: "1 == 2", expectedConstants: []interface{}{1, 2}, expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpEqual),
				code.Make(code.OpPop),
			},
		},
		{
			input: "1 != 2", expectedConstants: []interface{}{1, 2}, expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpNotEqual),
				code.Make(code.OpPop),
			},
		},
		{
			input: "true == false", expectedConstants: []interface{}{}, expectedInstructions: []code.Instructions{
				code.Make(code.OpTrue),
				code.Make(code.OpFalse),
				code.Make(code.OpEqual),
				code.Make(code.OpPop),
			},
		},
		{
			input: "true != false", expectedConstants: []interface{}{}, expectedInstructions: []code.Instructions{
				code.Make(code.OpTrue),
				code.Make(code.OpFalse),
				code.Make(code.OpNotEqual),
				code.Make(code.OpPop),
			},
		},
	}
	runCompilerTests(t, tests)
}

func TestIntegerArithmetic(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             `1 + 2;`,
			expectedConstants: []interface{}{1, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpAdd),
				code.Make(code.OpPop),
			},
		},
		{
			input: "1; 2", expectedConstants: []interface{}{1, 2}, expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpPop),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpPop),
			},
		},
		{
			input: "2 / 1", expectedConstants: []interface{}{2, 1}, expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpDiv),
				code.Make(code.OpPop),
			},
		},
		{
			input: "1 * 2", expectedConstants: []interface{}{1, 2}, expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpMul),
				code.Make(code.OpPop),
			},
		},
		{
			input: "1 - 2", expectedConstants: []interface{}{1, 2}, expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpSub),
				code.Make(code.OpPop),
			},
		},
	}
	runCompilerTests(t, tests)
}

func runCompilerTests(t *testing.T, tests []compilerTestCase) {
	t.Helper()

	for i, tt := range tests {
		program := parse(tt.input)
		compiler := New()
		err := compiler.Compile(program)
		if err != nil {
			t.Fatalf("test[%d]: compiler error: %s", i, err)
		}

		bytecode := compiler.Bytecode()
		err = testInstructions(tt.expectedInstructions, bytecode.Instructions)
		if err != nil {
			t.Errorf("test[%d]: testInstructions failed: %s", i, err)
		}
		err = testConstants(tt.expectedConstants, bytecode.Constants)
		if err != nil {
			t.Errorf("test[%d]: testConstants failed: %s", i, err)
		}
	}
}

func testInstructions(expected []code.Instructions, actual code.Instructions) error {
	concat := code.Instructions{}
	for _, instruction := range expected {
		concat = append(concat, instruction...)
	}

	if len(concat) != len(actual) {
		return fmt.Errorf("wrong instructions length.\nwant=%q\ngot=%q", concat, actual)
	}

	for i, expectedIns := range concat {
		if expectedIns != actual[i] {
			return fmt.Errorf("wrong instruction at %d.\nwant=%q\ngot =%q", i, concat, actual)
		}
	}
	return nil
}

func testConstants(expected []interface{}, actual []object.Object) error {
	if len(expected) != len(actual) {
		return fmt.Errorf("wrong number of constants. expected=%d, got=%d", len(expected), len(actual))
	}

	for i, expectedVal := range expected {
		switch constant := expectedVal.(type) {
		case int:
			err := testIntegerObject(actual[i], int64(constant))
			if err != nil {
				return fmt.Errorf("constant %d - testIntegerObject failed: %s", i, err)
			}
		case []code.Instructions:
			return fmt.Errorf(
				"constant %d - testInstructions failed: providing []code.Instructions is no longer supported. "+
					"Refactor tests to provide an expectedCompiledFunction instead.", i,
			)
		case expectedCompiledFunction:
			err := testInstructions(constant.instructions, actual[i].(*object.CompiledFunction).Instructions)
			if err != nil {
				return fmt.Errorf("constant %d - testInstructions failed: %s", i, err)
			}
			// compare numLocals and numParameters
			if constant.numLocals != actual[i].(*object.CompiledFunction).NumLocals {
				return fmt.Errorf("constant %d - NumLocals is wrong. want=%d, got=%d", i, constant.numLocals, actual[i].(*object.CompiledFunction).NumLocals)
			}
		default:
			return fmt.Errorf("unsupported expectedVal. got=%T (%+v)", expectedVal, expectedVal)
		}
	}

	return nil
}

func testIntegerObject(actual object.Object, expected int64) error {
	int, ok := actual.(*object.Integer)
	if !ok {
		return fmt.Errorf("actual is not Integer. got=%T (%+v)", actual, actual)
	}

	if int.Value != expected {
		return fmt.Errorf("int.Value is not %d. got=%d", expected, int.Value)
	}
	return nil
}

func parse(input string) *ast.Program {
	l := lexer.New(input)
	p := parser.New(l)
	return p.ParseProgram()
}

func TestPrefixExpressions(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "!true",
			expectedConstants: []interface{}{},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpTrue),
				code.Make(code.OpBang),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "-5",
			expectedConstants: []interface{}{5},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpMinus),
				code.Make(code.OpPop),
			},
		},
	}
	runCompilerTests(t, tests)
}

func TestIfExpressions(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "if (true) { 1; }; 3333;",
			expectedConstants: []interface{}{1, 3333},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpTrue),              // 0000
				code.Make(code.OpJumpNotTruthy, 10), // 0001
				code.Make(code.OpConstant, 0),       // 0004
				code.Make(code.OpJump, 11),          // 0007
				code.Make(code.OpNull),              // 0010
				code.Make(code.OpPop),               // 0011
				code.Make(code.OpConstant, 1),       // 0014
				code.Make(code.OpPop),               // 0017
			},
		},
		{
			input:             "if (true) { 1; } else { 2; }; 3333;",
			expectedConstants: []interface{}{1, 2, 3333},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpTrue),              // 0000
				code.Make(code.OpJumpNotTruthy, 10), // 0001
				code.Make(code.OpConstant, 0),       // 0004
				code.Make(code.OpJump, 13),          // 0007
				code.Make(code.OpConstant, 1),       // 0010
				code.Make(code.OpPop),               // 0013
				code.Make(code.OpConstant, 2),       // 0014
				code.Make(code.OpPop),               // 0017
			},
		},
	}
	runCompilerTests(t, tests)
}

func TestLetStatements(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "let x = 1;",
			expectedConstants: []interface{}{1},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0), // this means take what is on the stack, and assign it to symbol 0
			},
		},
		{
			input:             "let x = 1; x;",
			expectedConstants: []interface{}{1},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0), // this means take what is on the stack, and assign it to symbol 0
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "let x = 1; let y = x + x; y",
			expectedConstants: []interface{}{1},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0), // this means take what is on the stack, and assign it to symbol 0
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpAdd),
				code.Make(code.OpSetGlobal, 1),
				code.Make(code.OpGetGlobal, 1),
				code.Make(code.OpPop),
			},
		},
	}
	runCompilerTests(t, tests)
}

func TestFunctions(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: `fn() { let x = 1; let y = x + 1;  y * y }()`,
			expectedConstants: []interface{}{
				1,
				1,
				expectedCompiledFunction{
					instructions: []code.Instructions{
						code.Make(code.OpConstant, 0),
						code.Make(code.OpSetLocal, 0),
						code.Make(code.OpGetLocal, 0),
						code.Make(code.OpConstant, 1),
						code.Make(code.OpAdd),
						code.Make(code.OpSetLocal, 1),
						code.Make(code.OpGetLocal, 1),
						code.Make(code.OpGetLocal, 1),
						code.Make(code.OpMul),
						code.Make(code.OpReturnValue),
					},
					numLocals:     2,
					numParameters: 0,
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 2),
				code.Make(code.OpCall, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input: `fn() { return 5 + 10 }`,
			expectedConstants: []interface{}{
				5,
				10,
				expectedCompiledFunction{
					instructions: []code.Instructions{
						code.Make(code.OpConstant, 0),
						code.Make(code.OpConstant, 1),
						code.Make(code.OpAdd),
						code.Make(code.OpReturnValue),
					},
					numLocals:     0,
					numParameters: 0,
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 2),
				code.Make(code.OpPop),
			},
		},
		{
			input: `fn() { 5 + 10 }`, expectedConstants: []interface{}{
				5,
				10,
				expectedCompiledFunction{
					instructions: []code.Instructions{
						code.Make(code.OpConstant, 0),
						code.Make(code.OpConstant, 1),
						code.Make(code.OpAdd),
						code.Make(code.OpReturnValue),
					},
					numLocals:     0,
					numParameters: 0,
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 2),
				code.Make(code.OpPop),
			},
		},
		{
			input: `fn() { }`, expectedConstants: []interface{}{
				expectedCompiledFunction{
					instructions: []code.Instructions{
						code.Make(code.OpReturn),
					},
					numLocals:     0,
					numParameters: 0,
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input: `fn() { 1; 2 }`, expectedConstants: []interface{}{
				1,
				2,
				expectedCompiledFunction{
					instructions: []code.Instructions{
						code.Make(code.OpConstant, 0),
						code.Make(code.OpPop),
						code.Make(code.OpConstant, 1),
						code.Make(code.OpReturnValue),
					},
					numLocals:     0,
					numParameters: 0,
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 2),
				code.Make(code.OpPop),
			},
		},
		{
			input: `let a = fn() { }`, expectedConstants: []interface{}{
				expectedCompiledFunction{
					instructions: []code.Instructions{
						code.Make(code.OpReturn),
					},
					numLocals:     0,
					numParameters: 0,
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
			},
		},
	}
	runCompilerTests(t, tests)
}

func TestFunctionCalls(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: `fn() { 5 }()`,
			expectedConstants: []interface{}{
				5,
				expectedCompiledFunction{
					instructions: []code.Instructions{
						code.Make(code.OpConstant, 0),
						code.Make(code.OpReturnValue),
					},
					numLocals:     0,
					numParameters: 0,
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 1),
				code.Make(code.OpCall, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input: `let x = fn() { 5 }; x()`,
			expectedConstants: []interface{}{
				5,
				expectedCompiledFunction{
					instructions: []code.Instructions{
						code.Make(code.OpConstant, 0),
						code.Make(code.OpReturnValue),
					},
					numLocals:     0,
					numParameters: 0,
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 1),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpCall, 0),
				code.Make(code.OpPop),
			},
		},
	}
	runCompilerTests(t, tests)
}

func TestLetStatementScopes(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: `let a = fn() { let xyz = 1; xyz };`, expectedConstants: []interface{}{
				1,
				expectedCompiledFunction{
					instructions: []code.Instructions{
						code.Make(code.OpConstant, 0),
						code.Make(code.OpSetLocal, 0),
						code.Make(code.OpGetLocal, 0),
						code.Make(code.OpReturnValue),
					},
					numLocals:     1,
					numParameters: 0,
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 1),
				code.Make(code.OpSetGlobal, 0),
			},
		},
		{
			input: `
let num = 55;
fn() { num }
`,
			expectedConstants: []interface{}{
				55,
				expectedCompiledFunction{
					instructions: []code.Instructions{
						code.Make(code.OpGetGlobal, 0),
						code.Make(code.OpReturnValue),
					},
					numLocals:     0,
					numParameters: 0,
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpPop),
			},
		},
		{
			input: `
               fn() {
                   let num = 55;
num }
`,
			expectedConstants: []interface{}{
				55,
				expectedCompiledFunction{
					instructions: []code.Instructions{
						code.Make(code.OpConstant, 0),
						code.Make(code.OpSetLocal, 0),
						code.Make(code.OpGetLocal, 0),
						code.Make(code.OpReturnValue),
					},
					numLocals:     1,
					numParameters: 0,
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 1),
				code.Make(code.OpPop),
			},
		},
		{
			input: `
let firstFoobar = fn() { let foobar = 50; foobar; };
let secondFoobar = fn() { let foobar = 100; foobar; };
firstFoobar() + secondFoobar();
   `,
			expectedConstants: []interface{}{
				50,
				expectedCompiledFunction{
					instructions: []code.Instructions{
						code.Make(code.OpConstant, 0),
						code.Make(code.OpSetLocal, 0),
						code.Make(code.OpGetLocal, 0),
						code.Make(code.OpReturnValue),
					},
					numLocals:     1,
					numParameters: 0,
				},
				100,
				expectedCompiledFunction{
					instructions: []code.Instructions{
						code.Make(code.OpConstant, 2),
						code.Make(code.OpSetLocal, 0),
						code.Make(code.OpGetLocal, 0),
						code.Make(code.OpReturnValue),
					},
					numLocals:     1,
					numParameters: 0,
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 1),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpSetGlobal, 1),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpCall, 0),
				code.Make(code.OpGetGlobal, 1),
				code.Make(code.OpCall, 0),
				code.Make(code.OpAdd),
				code.Make(code.OpPop),
			},
		},
		{
			input: `
fn() {
                   let a = 55;
let b = 77;
a+b }
`,
			expectedConstants: []interface{}{
				55,
				77,
				expectedCompiledFunction{
					instructions: []code.Instructions{
						code.Make(code.OpConstant, 0),
						code.Make(code.OpSetLocal, 0),
						code.Make(code.OpConstant, 1),
						code.Make(code.OpSetLocal, 1),
						code.Make(code.OpGetLocal, 0),
						code.Make(code.OpGetLocal, 1),
						code.Make(code.OpAdd),
						code.Make(code.OpReturnValue),
					},
					numLocals:     2,
					numParameters: 0,
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 2),
				code.Make(code.OpPop),
			},
		},
	}
	runCompilerTests(t, tests)
}

func TestFunctionParameters(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: `fn(x, y) { return x + y };`,

			// the only constant we expect is the function itself
			expectedConstants: []interface{}{
				// get x, get y, add, return
				expectedCompiledFunction{
					instructions: []code.Instructions{
						code.Make(code.OpGetLocal, 0),
						code.Make(code.OpGetLocal, 1),
						code.Make(code.OpAdd),
						code.Make(code.OpReturnValue),
					},
					numLocals: 2,
				},
			},

			// we load the function onto the stack then pop it since we do nothing with it
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input: `fn(x, y) { let a = 1; return x + y + a };`,

			// the only constant we expect is the function itself
			expectedConstants: []interface{}{
				1,
				expectedCompiledFunction{
					instructions: []code.Instructions{
						code.Make(code.OpConstant, 0),
						code.Make(code.OpSetLocal, 2),
						code.Make(code.OpGetLocal, 0),
						code.Make(code.OpGetLocal, 1),
						code.Make(code.OpAdd),
						code.Make(code.OpGetLocal, 2),
						code.Make(code.OpAdd),
						code.Make(code.OpReturnValue),
					},
					numLocals: 3,
				},
			},

			// we load the function onto the stack then pop it since we do nothing with it
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 1),
				code.Make(code.OpPop),
			},
		},
		{
			input: `fn(x, y) { return x + y }(1, 2)`,

			// the first constant is the function
			// then 1 and 2
			expectedConstants: []interface{}{
				// get x, get y, add, return
				expectedCompiledFunction{
					instructions: []code.Instructions{
						code.Make(code.OpGetLocal, 0),
						code.Make(code.OpGetLocal, 1),
						code.Make(code.OpAdd),
						code.Make(code.OpReturnValue),
					},
					numLocals: 2,
				},
				1,
				2,
			},

			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpCall, 2),
				code.Make(code.OpPop),
			},
		},
	}
	runCompilerTests(t, tests)
}
