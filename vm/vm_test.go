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

func testStringObject(actual object.Object, expected string) error {
	string, ok := actual.(*object.String)
	if !ok {
		return fmt.Errorf("actual is not String. got=%T (%+v)", actual, actual)
	}

	if string.Value != expected {
		return fmt.Errorf("string.Value is not %q. got=%q", expected, string.Value)
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
		stackElem := vm.LastPoppedStackElement()
		testExpectedObject(t, stackElem, tt.expected)
	}
}

func testBooleanObject(actual object.Object, expected bool) error {
	boolean, ok := actual.(*object.Boolean)
	if !ok {
		return fmt.Errorf("actual is not Boolean. got=%T (%+v)", actual, actual)
	}

	if boolean.Value != expected {
		return fmt.Errorf("boolean.Value is not %t. got=%t", expected, boolean.Value)
	}
	return nil
}

func testExpectedObject(t *testing.T, actual object.Object, expected interface{}) {
	t.Helper()
	switch expected := expected.(type) {
	case int:
		err := testIntegerObject(actual, int64(expected))
		if err != nil {
			t.Errorf("testIntegerObject failed: %s", err)
		}
	case string:
		err := testStringObject(actual, expected)
		if err != nil {
			t.Errorf("testStringObject failed: %s", err)
		}
	case bool:
		err := testBooleanObject(actual, expected)
		if err != nil {
			t.Errorf("testBooleanObject failed: %s", err)
		}

	default:
		t.Errorf("unhandled test case in testExpectedObject: %T (%+v)", expected, expected)
	}
}

func TestIntegerArithmetic(t *testing.T) {
	tests := []vmTestCase{
		{`1 + 2`, 3},
		{`3 + 3`, 6},
		{`1 - 2`, -1},
		{"2 * 2;", 4},
		{"4 / 2", 2},
		{"50 / 2 * 2 + 10 - 5", 55},
		{"5 + 5 + 5 + 5 - 10", 10},
		{"2 * 2 * 2 * 2 * 2", 32},
		{"5 * 2 + 10", 20},
		{"5 + 2 * 10", 25},
		{"5 * (2 + 10)", 60},
		{"-1 + 1", 0},
		{"-5", -5},
		{"-10", -10},
		{"-50 + 100 + -50", 0},
		{"(5 + 10 * 2 + 15 / 3) * 2 + -10", 50},
	}
	runVmTests(t, tests)
}

func TestBooleanExpressions(t *testing.T) {
	tests := []vmTestCase{
		{`false`, false},
		{`true`, true},
		{"1 < 2", true},
		{"1 > 2", false},
		{"1 < 1", false},
		{"1 > 1", false},
		{"1 == 1", true},
		{"1 != 1", false},
		{"1 == 2", false},
		{"1 != 2", true},
		{"true == true", true},
		{"false == false", true},
		{"true == false", false},
		{"true != false", true},
		{"false != true", true},
		{"(1 < 2) == true", true},
		{"(1 < 2) == false", false},
		{"(1 > 2) == true", false},
		{"(1 > 2) == false", true},
		{"!true", false},
		{"!!true", true},
		{"!false", true},
		{"!!false", false},
		{"!!(1 < 2)", true},
		{"!((1 > 2) == false)", false},
		{"!true", false},
		{"!false", true},
		{"!!true", true},
		{"!!false", false},
	}
	runVmTests(t, tests)
}

func TestConditionals(t *testing.T) {
	tests := []vmTestCase{
		{"if (true) { 10 }", 10},
		{"if (true) { 10 } else { 20 }", 10},
		{"if (false) { 10 } else { 20 } ", 20},
		{"if (1 < 2) { 10 }", 10},
		{"if (1 < 2) { 10 } else { 20 }", 10},
		{"if (1 > 2) { 10 } else { 20 }", 20},
	}
	runVmTests(t, tests)
}
