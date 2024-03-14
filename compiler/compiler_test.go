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

	for _, tt := range tests {
		program := parse(tt.input)
		compiler := New()
		err := compiler.Compile(program)
		if err != nil {
			t.Fatalf("compiler error: %s", err)
		}

		bytecode := compiler.Bytecode()
		err = testInstructions(tt.expectedInstructions, bytecode.Instructions)
		if err != nil {
			t.Errorf("testInstructions failed: %s", err)
		}
		err = testConstants(tt.expectedConstants, bytecode.Constants)
		if err != nil {
			t.Errorf("testConstants failed: %s", err)
		}
	}
}

func testInstructions(expected []code.Instructions, actual code.Instructions) error {
	concat := code.Instructions{}
	for _, instruction := range expected {
		concat = append(concat, instruction...)
	}

	if len(concat) != len(actual) {
		return fmt.Errorf("wrong instructions length.\nwant=%q\ngot =%q", concat, actual)
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
