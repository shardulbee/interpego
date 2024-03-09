package vm

import (
	"fmt"
	"testing"

	"interpego/ast"
	"interpego/compiler"
	"interpego/lexer"
	"interpego/object"
	"interpego/parser"
)

func parse(input string) *ast.Program {
	l := lexer.New(input)
	p := parser.New(l)
	return p.ParseProgram()
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

type vmTestCase struct {
	input    string
	expected interface{}
}

func runVmTests(t *testing.T, tests []vmTestCase) {
	t.Helper()

	for _, tt := range tests {
		program := parse(tt.input)
		compiler := compiler.New()
		err := compiler.Compile(program)
		if err != nil {
			t.Fatalf("compiler error: %s", err)
		}

		vm := New(compiler.Bytecode())
		err = vm.Run()
		if err != nil {
			t.Fatalf("vm error: %s", err)
		}
		stackElem := vm.StackTop()
		testExpectedObject(t, stackElem, tt.expected)
	}
}

func testExpectedObject(t *testing.T, actual object.Object, expected interface{}) {
	t.Helper()
	switch expected := expected.(type) {
	case int:
		err := testIntegerObject(actual, int64(expected))
		if err != nil {
			t.Errorf("testIntegerObject failed: %s", err)
		}
	}
}

func TestIntegerArithmetic(t *testing.T) {
	tests := []vmTestCase{
		{`1`, 1},
		{`2`, 2},
		{`1 + 2`, 2}, // FIXME
	}
	runVmTests(t, tests)
}
