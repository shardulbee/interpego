package evaluator

import (
	"interpego/lexer"
	"interpego/object"
	"interpego/parser"
	"testing"
)

func TestEvalIntegerExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"5", 5},
		{"10", 10},
		{"-5", -5},
		{"-10", -10},
		{"5 + 5", 10},
		{"5 / 5", 1},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testIntegerObject(t, evaluated, tt.expected)
	}
}

func TestEvalBooleanExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"true", true},
		{"false", false},
		{"!true", false},
		{"!false", true},
		{"!!true", true},
		{"!!false", false},
		{"!!!!!!!!!true", false},
	}

	for i, tt := range tests {
		t.Logf("Looking at test case: %d", i)
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.expected)
	}
}

func testEval(input string) object.Object {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	return Eval(program)
}

func testBooleanObject(t *testing.T, actual object.Object, expected bool) bool {
	boolean, ok := actual.(*object.Boolean)
	if !ok {
		t.Errorf("actual is not Boolean. got=%T (%+v)", actual, actual)
		return false
	}

	if boolean.Value != expected {
		t.Errorf("boolean.Value is not %t. got=%t", expected, boolean.Value)
		return false
	}
	return true
}

func testIntegerObject(t *testing.T, actual object.Object, expected int64) bool {
	int, ok := actual.(*object.Integer)
	if !ok {
		t.Errorf("actual is not Integer. got=%T (%+v)", actual, actual)
		return false
	}

	if int.Value != expected {
		t.Errorf("int.Value is not %d. got=%d", expected, int.Value)
		return false
	}
	return true
}
