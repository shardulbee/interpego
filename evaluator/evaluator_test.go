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
		{"5 + 5 + 5 + 5 - 10", 10},
		{"2 * 2 * 2 * 2 * 2", 32},
		{"-50 + 100 + -50", 0},
		{"5 * 2 + 10", 20},
		{"5 + 2 * 10", 25},
		{"20 + 2 * -10", 0},
		{"50 / 2 * 2 + 10", 60},
		{"2 * (5 + 10)", 30},
		{"3 * 3 * 3 + 10", 37},
		{"3 * (3 * 3) + 10", 37},
		{"(5 + 10 * 2 + 15 / 3) * 2 + -10", 50},
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
		{"null == true", false},
		{"null == 1", false},
		{"null == null", true},
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

func TestReturnStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"return 10;", 10},
		{"return 10; 9;", 10},
		{"return 2 * 5; 9;", 10},
		{"9; return 2 * 5; 9;", 10},
	}
	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testIntegerObject(t, evaluated, tt.expected)
	}
}

func TestIfElseExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"if (true) { 10 }", 10},
		{"if (false) { 10 }", nil},
		{"if (1) { 10 }", 10},
		{"if (1 < 2) { 10 }", 10},
		{"if (1 > 2) { 10 }", nil},
		{"if (1 > 2) { 10 } else { 20 }", 20},
		{"if (1 < 2) { 10 } else { 20 }", 10},
		{`
if (10 > 1) {
	if (10 > 1) {
		return 10;
	}
  return 1;
}
			`, 10},
	}
	for i, tt := range tests {
		t.Logf("Looking at test case: %d", i)
		actual := testEval(tt.input)
		int, ok := tt.expected.(int)
		if ok {
			testIntegerObject(t, actual, int64(int))
		} else {
			testNullObject(t, actual)
		}
	}
}

func testNullObject(t *testing.T, actual object.Object) bool {
	if actual != NULL {
		t.Errorf("expected actual to have type object.Null. got=%T", actual)
		return false
	}
	return true
}
