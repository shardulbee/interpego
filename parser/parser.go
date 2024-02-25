package parser

import (
	"fmt"
	"interpego/ast"
	"interpego/lexer"
	"interpego/token"
	"strconv"
)

const (
	_ int = iota
	LOWEST
	EQUALS
	LESSGREATER
	SUM
	PRODUCT
	PREFIX
	CALL
)

var precedences = map[token.TokenType]int{
	token.EQ:       EQUALS,
	token.NOT_EQ:   EQUALS,
	token.LT:       LESSGREATER,
	token.GT:       LESSGREATER,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.SLASH:    PRODUCT,
	token.ASTERISK: PRODUCT,
	token.LPAREN:   CALL,
}

type Error string

type Parser struct {
	lexer  *lexer.Lexer
	errors []string

	curToken token.Token
	// need this to look ahead to see if an expression is complete for example
	// let a = 5; <--- need to know if 5 is followed by PLUS for example to know that the RHS is an arithmetic
	// expression or SEMICOLON to know if it is EOL
	peekToken token.Token

	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

func New(lexer *lexer.Lexer) *Parser {
	p := Parser{lexer: lexer, errors: []string{}}

	p.nextToken()
	p.nextToken()

	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.infixParseFns = make(map[token.TokenType]infixParseFn)

	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.INT, p.parseIntegerLiteral)
	p.registerPrefix(token.TRUE, p.parseBoolean)
	p.registerPrefix(token.FALSE, p.parseBoolean)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.BANG, p.parsePrefixExpression)
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)
	p.registerPrefix(token.IF, p.parseIfExpression)
	p.registerPrefix(token.FUNCTION, p.parseFunctionLiteral)
	p.registerInfix(token.LPAREN, p.parseCallExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)

	return &p
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("expected next token to be %q, got %q instead", t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.lexer.NextToken()
}

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for !p.curTokenIs(token.EOF) {
		statement := p.parseStatement()
		if statement != nil {
			program.Statements = append(program.Statements, statement)
		}

		p.nextToken()
	}
	return program
}

// at the end of execution of this function, curToken should be a semicolon
func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.LET:
		return p.parseLetStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := ast.ExpressionStatement{Token: p.curToken}
	stmt.Expression = p.parseExpression(LOWEST)
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return &stmt
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := ast.ReturnStatement{Token: p.curToken}

	p.nextToken()
	stmt.ReturnValue = p.parseExpression(LOWEST)
	p.nextToken()

	return &stmt
}

func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) parseLetStatement() *ast.LetStatement {
	statement := ast.LetStatement{Token: p.curToken}

	if !p.expectPeek(token.IDENT) {
		return nil
	}
	statement.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.ASSIGN) {
		return nil
	}
	p.nextToken()
	statement.Value = p.parseExpression(LOWEST)
	p.nextToken()

	return &statement
}

// 1 == 1 > 2

func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.curToken.Type]

	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix()

	for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}

		p.nextToken()
		leftExp = infix(leftExp)
	}

	return leftExp
}

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: p.curToken}

	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}
	lit.Value = value
	return lit
}

func (p *Parser) parseBoolean() ast.Expression {
	return &ast.BooleanLiteral{Token: p.curToken, Value: p.curTokenIs(token.TRUE)}
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	exp := &ast.PrefixExpression{Token: p.curToken, Operator: p.curToken.Literal}
	p.nextToken()
	exp.Right = p.parseExpression(PREFIX)
	return exp
}

func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("no prefix parse fn found for %q", t)
	p.errors = append(p.errors, msg)
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	exp := &ast.InfixExpression{Token: p.curToken, Operator: p.curToken.Literal, Left: left}
	precedence := p.curPrecedence()
	p.nextToken()
	exp.Right = p.parseExpression(precedence)
	return exp
}

func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	} else {
		p.peekError(t)
		return false
	}
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()
	exp := p.parseExpression(LOWEST)
	if !p.expectPeek(token.RPAREN) {
		return nil
	}
	return exp
}

func (p *Parser) parseIfExpression() ast.Expression {
	ifExp := &ast.IfExpression{
		Token: p.curToken,
	}
	p.nextToken()
	ifExp.Condition = p.parseExpression(LOWEST)

	if !p.peekTokenIs(token.LBRACE) {
		p.peekError(token.LBRACE)
		return nil
	}
	p.nextToken()
	ifExp.Consequence = p.parseBlockStatement()

	if p.peekTokenIs(token.ELSE) {
		p.nextToken()
		if !p.peekTokenIs(token.LBRACE) {
			p.peekError(token.LBRACE)
			return nil
		}
		p.nextToken()
		ifExp.Alternative = p.parseBlockStatement()
	}

	return ifExp
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	stmt := &ast.BlockStatement{Token: p.curToken, Statements: []ast.Statement{}}
	p.nextToken()

	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		statement := p.parseStatement()
		if statement != nil {
			stmt.Statements = append(stmt.Statements, statement)
		}

		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseFunctionLiteral() ast.Expression {
	fn := &ast.FunctionLiteral{Token: p.curToken}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}
	fn.Parameters = p.parseFunctionParameters()

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	fn.FunctionBody = p.parseBlockStatement()

	return fn
}

func (p *Parser) parseFunctionParameters() []*ast.Identifier {
	params := []*ast.Identifier{}
	for !p.curTokenIs(token.RPAREN) && !p.curTokenIs(token.EOF) {
		if p.curTokenIs(token.IDENT) {
			params = append(params, &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal})
		}
		p.nextToken()
	}

	if !p.curTokenIs(token.RPAREN) {
		msg := fmt.Sprintf("expected RPAREN, got=%q", p.curToken.Type)
		p.errors = append(p.errors, msg)
		return nil
	}

	return params
}

func (p *Parser) parseCallExpression(left ast.Expression) ast.Expression {
	call := &ast.CallExpression{Token: p.curToken, Function: left}
	p.nextToken()

	args := []ast.Expression{}
	for !p.curTokenIs(token.RPAREN) && !p.curTokenIs(token.EOF) {
		args = append(args, p.parseExpression(LOWEST))
		if p.peekTokenIs(token.COMMA) {
			p.nextToken()
		}
		p.nextToken()
	}

	call.Arguments = args
	return call
}
