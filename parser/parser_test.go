package parser

import (
	"fmt"
	"interpego/ast"
	"interpego/lexer"
	"testing"
)

func TestReturnStatements(t *testing.T) {
	input := `
return 1;
return 10;
return add(2, 3);
`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)
	if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	}

	if len(program.Statements) != 3 {
		t.Fatalf("Expected 3 statements, found %d", len(program.Statements))
	}
	for _, stmt := range program.Statements {
		returnStmt, ok := stmt.(*ast.ReturnStatement)
		if !ok {
			t.Errorf("expected stmt to be *ast.ReturnStmt. got=%T", stmt)
			continue
		}
		if returnStmt.TokenLiteral() != "return" {
			t.Errorf("expected returnStmt.TokenLiteral() to be 'return', got=%q", returnStmt.TokenLiteral())
		}
	}
}

func TestLetStatements(t *testing.T) {
	input := `
let x = 5;
let y = 10;
let foobar= 83838383;
let more = 3 + 3;
`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)
	if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	}
	tests := []struct {
		expectedLiteral string
	}{
		{"x"},
		{"y"},
		{"foobar"},
		{"more"},
	}
	if len(program.Statements) != len(tests) {
		t.Fatalf("Expected %d statements, found %d", len(tests), len(program.Statements))
	}
	for i, tt := range tests {
		statement := program.Statements[i]
		if !testLetStatement(t, statement, tt.expectedLiteral) {
			return
		}
	}
}

func TestParsingInfixExpressions(t *testing.T) {
	tests := []struct {
		input      string
		leftValue  interface{}
		operator   string
		rightValue interface{}
	}{
		{"5 + 5;", 5, "+", 5},
		{"5 - 5;", 5, "-", 5},
		{"5 * 5;", 5, "*", 5},
		{"5 / 5;", 5, "/", 5},
		{"5 > 5;", 5, ">", 5},
		{"5 < 5;", 5, "<", 5},
		{"5 == 5;", 5, "==", 5},
		{"5 != 5;", 5, "!=", 5},
		{"true == true", true, "==", true},
		{"true != false", true, "!=", false},
		{"false == false", false, "==", false},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program does not have the right amount of statements. expected 1, got %d", len(program.Statements))
		}

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("program.Statements[0] is not an ast.ExpressionStatement. got=%T", program.Statements[0])
		}

		exp, ok := stmt.Expression.(*ast.InfixExpression)
		if !ok {
			t.Fatalf("stmt is not ast.InfixExpression. got=%T", stmt.Expression)
		}
		testInfixExpression(t, exp, tt.leftValue, tt.operator, tt.rightValue)
	}
}

func TestIdentifierExpression(t *testing.T) {
	input := `foobar;`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program does not have the right amount of statements. expected 1, got %d", len(program.Statements))
	}
	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not an ast.ExpressionStatement. got=%T", program.Statements[0])
	}

	ident, ok := stmt.Expression.(*ast.Identifier)
	if !ok {
		t.Fatalf("exp not *ast.Identifier. got=%T", stmt.Expression)
	}

	if ident.Value != "foobar" {
		t.Errorf("ident.Value not %s. got=%s", "foobar", ident.Value)
	}
	if ident.TokenLiteral() != "foobar" {
		t.Errorf("ident.TokenLiteral() not %s. got=%s", "foobar", ident.TokenLiteral())
	}
}

func TestIntegerLiteralExpression(t *testing.T) {
	input := `5;`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program does not have the right amount of statements. expected 1, got %d", len(program.Statements))
	}
	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not an ast.ExpressionStatement. got=%T", program.Statements[0])
	}

	ident, ok := stmt.Expression.(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("exp not *ast.Identifier. got=%T", stmt.Expression)
	}

	if ident.Value != 5 {
		t.Errorf("ident.Value not %d. got=%d", 5, ident.Value)
	}
	if ident.TokenLiteral() != "5" {
		t.Errorf("ident.TokenLiteral() not %s. got=%s", "5", ident.TokenLiteral())
	}
}

func TestParsingPrefixExpressions(t *testing.T) {
	tests := []struct {
		input        string
		operator     string
		integerValue interface{}
	}{
		{"!5;", "!", 5},
		{"-15;", "-", 15},
		{"!true;", "!", true},
		{"!false;", "!", false},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program does not have the right amount of statements. expected 1, got %d", len(program.Statements))
		}

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("program.Statements[0] is not an ast.ExpressionStatement. got=%T", program.Statements[0])
		}

		exp, ok := stmt.Expression.(*ast.PrefixExpression)
		if !ok {
			t.Fatalf("stmt is not ast.PrefixExpression. got=%T", stmt.Expression)
		}

		if exp.Operator != tt.operator {
			t.Fatalf("exp.Operator is not %q. got=%q", tt.operator, exp.Operator)
		}

		if !testLiteralExpression(t, exp.Right, tt.integerValue) {
			return
		}
	}
}

func TestOperatorPrecedeceParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			"-a * b",
			"((-a) * b)",
		},
		{
			"!-a",
			"(!(-a))",
		},
		{
			"a + b + c",
			"((a + b) + c)",
		},
		{
			"a + b - c",
			"((a + b) - c)",
		},
		{
			"a * b * c",
			"((a * b) * c)",
		},
		{
			"a * b / c",
			"((a * b) / c)",
		},
		{
			"a + b / c",
			"(a + (b / c))",
		},
		{
			"a + b * c + d / e - f",
			"(((a + (b * c)) + (d / e)) - f)",
		},
		{
			"3 + 4; -5 * 5",
			"(3 + 4)((-5) * 5)",
		},
		{
			"5 > 4 == 3 < 4",
			"((5 > 4) == (3 < 4))",
		},
		{
			"5 < 4 != 3 > 4",
			"((5 < 4) != (3 > 4))",
		},
		{
			"3 + 4 * 5 == 3 * 1 + 4 * 5",
			"((3 + (4 * 5)) == ((3 * 1) + (4 * 5)))",
		},
		{
			"1 == 1 > 2",
			"(1 == (1 > 2))",
		},
		{
			"2 + 2 * 3",
			"(2 + (2 * 3))",
		},
		{
			"2 * (2 + 2 * 3)",
			"(2 * (2 + (2 * 3)))",
		},
		{
			"2 * -(2 + 2 * 3)",
			"(2 * (-(2 + (2 * 3))))",
		},
		{
			"(2 + 2 * 3)",
			"(2 + (2 * 3))",
		},
		{
			"2 + 2 * 3",
			"(2 + (2 * 3))",
		},
		{
			"2 / (5 + 5)",
			"(2 / (5 + 5))",
		},
		{
			"-(5 + 5)",
			"(-(5 + 5))",
		},
		{
			"!(true == true)",
			"(!(true == true))",
		},
		{
			"(5 + 5) * 2",
			"((5 + 5) * 2)",
		},
		{
			"1 + (2 + 3) + 4",
			"((1 + (2 + 3)) + 4)",
		},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		actual := program.String()
		if tt.expected != actual {
			t.Errorf("expected=%q, got=%q", tt.expected, actual)
		}
	}
}

func testReturnStatement(t *testing.T, statement ast.Statement) bool {
	if statement.TokenLiteral() != "return" {
		t.Errorf("statement.TokenLiteral not 'return'. got %q", statement.TokenLiteral())
		return false
	}
	return true
}

func testLetStatement(t *testing.T, s ast.Statement, expectedIdentifier string) bool {
	if s.TokenLiteral() != "let" {
		t.Errorf("s.TokenLiteral not 'let'. got %q", s.TokenLiteral())
		return false
	}
	letStatement, ok := s.(*ast.LetStatement)
	if !ok {
		t.Errorf("s not *ast.LetStatement. got=%T", s)
		return false
	}

	if letStatement.Name.Value != expectedIdentifier {
		t.Errorf("letStatement.Name.Value not '%s'. got=%q", expectedIdentifier, letStatement.Name.Value)
		return false
	}

	if letStatement.Name.TokenLiteral() != expectedIdentifier {
		t.Errorf("letStatement.Name.TokenLiteral() not '%s'. got=%s", expectedIdentifier, letStatement.Name.TokenLiteral())
		return false
	}
	return true
}

func checkParserErrors(t *testing.T, p *Parser) {
	if len(p.Errors()) > 0 {
		t.Errorf("parser has %d errors", len(p.Errors()))
		for _, err := range p.Errors() {
			t.Errorf("parser error: %s", err)
		}
		t.FailNow()
	}
}

func testIntegerLiteral(t *testing.T, expression ast.Expression, expected int64) bool {
	il, ok := expression.(*ast.IntegerLiteral)
	if !ok {
		t.Errorf("expression is not an ast.IntegerLiteral. got=%T", expression)
		return false
	}

	if il.Value != expected {
		t.Errorf("il.Value is not %d. got=%d", expected, il.Value)
		return false
	}
	return true
}

func testIdentifier(t *testing.T, expression ast.Expression, expected string) bool {
	ident, ok := expression.(*ast.Identifier)
	if !ok {
		t.Errorf("expression is not an ast.Identifier. got=%T", expression)
		return false
	}

	if ident.Value != expected {
		t.Errorf("ident.Value is not %q. got=%q", expected, ident.Value)
		return false
	}

	if ident.TokenLiteral() != expected {
		t.Errorf("ident.TokenLiteral() is not %q. got=%q", expected, ident.TokenLiteral())
		return false
	}

	return true
}

func testBooleanLiteral(t *testing.T, expression ast.Expression, expected bool) bool {
	bl, ok := expression.(*ast.BooleanLiteral)
	if !ok {
		t.Errorf("expression is not an ast.BooleanLiteral. got=%T", expression)
		return false
	}

	if bl.Value != expected {
		t.Errorf("bl.Value is not %t. got=%t", expected, bl.Value)
		return false
	}

	if bl.TokenLiteral() != fmt.Sprintf("%t", expected) {
		t.Errorf("bl.TokenLiteral() is not %q. got=%q", fmt.Sprintf("%t", expected), bl.TokenLiteral())
		return false
	}

	return true
}

func testLiteralExpression(t *testing.T, exp ast.Expression, expected interface{}) bool {
	switch v := expected.(type) {
	case int:
		return testIntegerLiteral(t, exp, int64(v))
	case int64:
		return testIntegerLiteral(t, exp, v)
	case string:
		return testIdentifier(t, exp, v)
	case bool:
		return testBooleanLiteral(t, exp, v)
	}
	t.Errorf("type of exp not handled. got=%T", exp)
	return false
}

func testInfixExpression(t *testing.T, exp ast.Expression, left interface{}, operator string, right interface{}) bool {
	infix, ok := exp.(*ast.InfixExpression)
	if !ok {
		t.Errorf("exp is not an ast.InfixExpression. got=%T", exp)
		return false
	}

	if !testLiteralExpression(t, infix.Left, left) {
		return false
	}
	if infix.Operator != operator {
		t.Errorf("infix.Operator is not %s. got=%s", operator, infix.Operator)
		return false
	}
	if !testLiteralExpression(t, infix.Right, right) {
		return false
	}

	return true
}

func TestParsingBooleanExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"true;", true},
		{"false;", false},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program does not have the right amount of statements. expected 1, got %d", len(program.Statements))
		}
		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("program.Statements[0] is not an ast.ExpressionStatement. got=%T", program.Statements[0])
		}

		testLiteralExpression(t, stmt.Expression, tt.expected)
	}
}

func TestParsingIfExpressions(t *testing.T) {
	input := `if (x > y) { x } else { y };`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program does not have the right amount of statements. expected 1, got %d", len(program.Statements))
	}
	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not an ast.ExpressionStatement. got=%T", program.Statements[0])
	}

	ifexp, ok := stmt.Expression.(*ast.IfExpression)
	if !ok {
		t.Fatalf("stmt.Expression is not an ast.IfExpression. got=%T", stmt.Expression)
	}

	if !testInfixExpression(t, ifexp.Condition, "x", ">", "y") {
		return
	}

	if len(ifexp.Consequence.Statements) != 1 {
		t.Fatalf("ifexp.Consequence does not have the right amount of statements. expected 1, got %d", len(ifexp.Consequence.Statements))
	}
	consequenceStmt, ok := ifexp.Consequence.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("ifexp.Consequence.Statements[0] is not an ast.ExpressionStatement. got=%T", program.Statements[0])
	}
	if !testIdentifier(t, consequenceStmt.Expression, "x") {
		return
	}

	if len(ifexp.Alternative.Statements) != 1 {
		t.Fatalf("ifexp.Alternative does not have the right amount of statements. expected 1, got %d", len(ifexp.Alternative.Statements))
	}
	alternativeStmt, ok := ifexp.Alternative.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("ifexp.Alternative.Statements[0] is not an ast.ExpressionStatement. got=%T", program.Statements[0])
	}
	if !testIdentifier(t, alternativeStmt.Expression, "y") {
		return
	}
}

func TestParsingBoolean(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"true;", true},
		{"false;", false},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program does not have the right amount of statements. expected 1, got %d", len(program.Statements))
		}
		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("program.Statements[0] is not an ast.ExpressionStatement. got=%T", program.Statements[0])
		}

		testLiteralExpression(t, stmt.Expression, tt.expected)
	}
}

func TestParsingIfExpressionsNoAlternative(t *testing.T) {
	input := `if (x > y) { x };`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program does not have the right amount of statements. expected 1, got %d", len(program.Statements))
	}
	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not an ast.ExpressionStatement. got=%T", program.Statements[0])
	}

	ifexp, ok := stmt.Expression.(*ast.IfExpression)
	if !ok {
		t.Fatalf("stmt.Expression is not an ast.IfExpression. got=%T", stmt.Expression)
	}

	if !testInfixExpression(t, ifexp.Condition, "x", ">", "y") {
		return
	}

	if len(ifexp.Consequence.Statements) != 1 {
		t.Fatalf("ifexp.Consequence does not have the right amount of statements. expected 1, got %d", len(ifexp.Consequence.Statements))
	}
	consequenceStmt, ok := ifexp.Consequence.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("ifexp.Consequence.Statements[0] is not an ast.ExpressionStatement. got=%T", program.Statements[0])
	}
	if !testIdentifier(t, consequenceStmt.Expression, "x") {
		return
	}

	if ifexp.Alternative != nil {
		t.Fatal("ifexp.Alternative was expected to be nil but it is not")
	}
}

func TestParsingFunctionLiteral(t *testing.T) {
	input := `
	fn(x, y) {
		x + y;
	};
	`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program does not have the right amount of statements. expected 1, got %d", len(program.Statements))
	}
	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not an ast.ExpressionStatement. got=%T", program.Statements[0])
	}

	ifexp, ok := stmt.Expression.(*ast.FunctionLiteral)
	if !ok {
		t.Fatalf("stmt.Expression is not an ast.FunctionLiteral. got=%T", stmt.Expression)
	}

	if len(ifexp.Parameters) != 2 {
		t.Fatalf("expected ifexp.Parameters to contain 2 identifiers. got=%d", len(ifexp.Parameters))
	}
	if !testIdentifier(t, ifexp.Parameters[0], "x") {
		return
	}
	if !testIdentifier(t, ifexp.Parameters[1], "y") {
		return
	}

	if len(ifexp.FunctionBody.Statements) != 1 {
		t.Fatalf("expected ifexp.FunctionBody.Statements to contain 1 statements. got=%d", len(ifexp.FunctionBody.Statements))
	}
	exp, ok := ifexp.FunctionBody.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("expected ifexp.FunctionBody.Statements[0] to be an ast.ExpressionStatement. got=%T", ifexp.FunctionBody.Statements[0])
	}
	testInfixExpression(t, exp.Expression, "x", "+", "y")
}
func TestFunctionParameterParsing(t *testing.T) {
	tests := []struct {
		input          string
		expectedParams []string
	}{
		{input: "fn() {};", expectedParams: []string{}},
		{input: "fn(x) {};", expectedParams: []string{"x"}},
		{input: "fn(x, y, z) {};", expectedParams: []string{"x", "y", "z"}},
	}
	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)
		stmt := program.Statements[0].(*ast.ExpressionStatement)
		function := stmt.Expression.(*ast.FunctionLiteral)
		if len(function.Parameters) != len(tt.expectedParams) {
			t.Errorf("length parameters wrong. want %d, got=%d\n",
				len(tt.expectedParams), len(function.Parameters))
		}
		for i, ident := range tt.expectedParams {
			testLiteralExpression(t, function.Parameters[i], ident)
		}
	}
}
