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

func TestEvalStringExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"thing"`, "thing"},
		{`"hello" + " " + "world"`, "hello world"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testStringObject(t, evaluated, tt.expected)
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
		{`"hello" == "hello"`, true},
		{`"hello" == "goodbye"`, false},
		{`"hello" != "goodbye"`, true},
		{`"hello" != "hello"`, false},
		{`"a" < "b"`, true},
		{`"b" > "a"`, true},
		{`"a" > "b"`, false},
		{`"b" < "a"`, false},
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
	env := object.NewEnvironment()
	return Eval(program, env)
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

func testStringObject(t *testing.T, actual object.Object, expected string) bool {
	string, ok := actual.(*object.String)
	if !ok {
		t.Errorf("actual is not String. got=%T (%+v)", actual, actual)
		return false
	}

	if string.Value != expected {
		t.Errorf("string.Value is not %q. got=%q", expected, string.Value)
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

func TestErrorHandling(t *testing.T) {
	tests := []struct {
		input           string
		expectedMessage string
	}{
		{"5 + true;", "type mismatch: INTEGER + BOOLEAN"},
		{"5 + true; 5;", "type mismatch: INTEGER + BOOLEAN"},
		{"-true", "unknown operator: -BOOLEAN"},
		{"true + false;", "unknown operator: BOOLEAN + BOOLEAN"},
		{"if (10 > 1) { true + false; }", "unknown operator: BOOLEAN + BOOLEAN"},
		{"!1", "unknown operator: !INTEGER"},
		{`"Hello" - "World"`, "unknown operator: STRING - STRING"},
		{
			"foobar",
			"unknown identifier: foobar",
		},
	}
	for i, tt := range tests {
		t.Logf("Looking at test case: %d", i)
		result := testEval(tt.input)
		errorObj, ok := result.(*object.Error)
		if !ok {
			t.Fatalf("result is not an object.Error. got=%T (%+v)", result, result)
		}

		if errorObj.Message != tt.expectedMessage {
			t.Fatalf("errorObj.Message is not %q. got=%q", tt.expectedMessage, errorObj.Message)
		}
	}

}

func TestLetStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"let a = 5; a;", 5},
		{"let a = 5 * 5; a;", 25},
		{"let a = 5; let b = a; b;", 5},
		{"let a = 5; let b = a; let c = a + b + 5; c;", 15},
		{"let a = if (true) { return 10 } else { 5 }; a;", 10},
	}
	for i, tt := range tests {
		t.Logf("Looking at test case: %d", i)
		result := testEval(tt.input)
		testIntegerObject(t, result, tt.expected)
	}
}

func TestFunctionObject(t *testing.T) {
	input := `fn(x) { x + 2; };`
	result := testEval(input)
	fn, ok := result.(*object.Function)
	if !ok {
		t.Fatalf("result is not an object.Function. got=%T", result)
	}

	if len(fn.Params) != 1 {
		t.Fatalf("wrong number of arguments. expected=1, got=%d", len(fn.Params))
	}
	if fn.Params[0].Value != "x" {
		t.Fatalf("expected fn.Params[0] to be 'x'. got=%q", fn.Params[0].Value)
	}

	if len(fn.Body.Statements) != 1 {
		t.Fatalf("wrong number of statements in function body. expected=1, got=%d", len(fn.Body.Statements))
	}
	expectedBody := "(x + 2)"
	if fn.Body.String() != expectedBody {
		t.Fatalf("body is not %q. got=%q", expectedBody, fn.Body.String())
	}
}

func TestFunctionApplication(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"let identity = fn(x) { x; }; identity(5);", 5},
		{"let identity = fn(x) { return x; }; identity(5);", 5},
		{"let double = fn(x) { x * 2; }; double(5);", 10},
		{"let add = fn(x, y) { x + y; }; add(5, 5);", 10},
		{"let add = fn(x, y) { x + y; }; add(5 + 5, add(5, 5));", 20},
		{"fn(x) { x; }(5)", 5},
		{"fn(x) { x; return 2; }(1)", 2},
		{"fn(x) { x; return 2; x; }(1)", 2},
		{"fn(x) { x; return 2; x; }(1); return 20", 20},
	}
	for _, tt := range tests {
		testIntegerObject(t, testEval(tt.input), tt.expected)
	}
}

func TestClosures(t *testing.T) {
	input := `
   let newAdder = fn(x) {
     fn(y) { x + y };
};
   let addTwo = newAdder(2);
   addTwo(2);`
	testIntegerObject(t, testEval(input), 4)

	input = `
		let a = 1;
		let addA = fn(x) { x + a };
		addA(3);
	`
	testIntegerObject(t, testEval(input), 4)
}

func TestBuiltinFunctions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{`len("")`, 0},
		{`len("four")`, 4},
		{`len("hello world")`, 11},
		{`len(1)`, "argument to `len` not supported, got INTEGER"},
		{`len("one", "two")`, "wrong number of arguments. got=2, want=1"},
	}
	for _, tt := range tests {
		evaluated := testEval(tt.input)
		switch expected := tt.expected.(type) {
		case int:
			testIntegerObject(t, evaluated, int64(expected))
		case string:
			errObj, ok := evaluated.(*object.Error)
			if !ok {
				t.Errorf("object is not Error. got=%T (%+v)",
					evaluated, evaluated)
				continue
			}
			if errObj.Message != expected {
				t.Errorf("wrong error message. expected=%q, got=%q",
					expected, errObj.Message)
			}
		}
	}
}
