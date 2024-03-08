package ast

import (
	"bytes"
	"fmt"

	"interpego/token"
)

type Node interface {
	TokenLiteral() string
	String() string
}

type Statement interface {
	Node
	statementNode()
}

type Expression interface {
	Node
	expressionNode()
}

type LetStatement struct {
	Token token.Token // this is the LET token
	Name  *Identifier
	Value Expression
}

func (ls *LetStatement) statementNode()       {}
func (ls *LetStatement) TokenLiteral() string { return ls.Token.Literal }
func (ls *LetStatement) String() string {
	var out bytes.Buffer
	out.WriteString(ls.TokenLiteral() + " ")
	out.WriteString(ls.Name.String())
	out.WriteString(" " + token.ASSIGN + " ")

	if ls.Value != nil {
		out.WriteString(ls.Value.String())
	}
	return out.String()
}

type ReturnStatement struct {
	Token       token.Token // this is the RETURN token
	ReturnValue Expression
}

func (rs *ReturnStatement) statementNode()       {}
func (rs *ReturnStatement) TokenLiteral() string { return rs.Token.Literal }
func (rs *ReturnStatement) String() string {
	var out bytes.Buffer
	out.WriteString(rs.TokenLiteral() + " ")
	if rs.ReturnValue != nil {
		out.WriteString(rs.ReturnValue.String())
	}
	out.WriteString(";")
	return out.String()
}

type ExpressionStatement struct {
	Token      token.Token // this is the first token in the expression
	Expression Expression
}

func (es *ExpressionStatement) statementNode()       {}
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}
	return ""
}

type Identifier struct {
	Token token.Token
	Value string
}

// identifier is an expression because you can do things
// like let a = b; where b is an expression
func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }
func (i *Identifier) String() string {
	return i.Value
}

type IntegerLiteral struct {
	Token token.Token
	Value int64
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }
func (il *IntegerLiteral) String() string       { return il.Token.Literal }

type StringLiteral struct {
	Token token.Token
	Value string
}

func (sl *StringLiteral) expressionNode()      {}
func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Literal }
func (sl *StringLiteral) String() string       { return sl.Token.Literal }

type BooleanLiteral struct {
	Token token.Token
	Value bool
}

func (bl *BooleanLiteral) expressionNode()      {}
func (bl *BooleanLiteral) TokenLiteral() string { return bl.Token.Literal }
func (bl *BooleanLiteral) String() string       { return bl.Token.Literal }

type ArrayLiteral struct {
	Elements []Expression
	Token    token.Token // the '[' token
}

func (al *ArrayLiteral) expressionNode()      {}
func (al *ArrayLiteral) TokenLiteral() string { return al.Token.Literal }
func (al *ArrayLiteral) String() string {
	var out bytes.Buffer

	out.WriteString("[")
	for i, exp := range al.Elements {
		out.WriteString(exp.String())
		if i != len(al.Elements)-1 {
			out.WriteString(", ")
		}
	}
	out.WriteString("]")
	return out.String()
}

type PrefixExpression struct {
	Token    token.Token // the prefix token i.e. !, -
	Operator string
	Right    Expression
}

func (pe *PrefixExpression) expressionNode()      {}
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Literal }
func (pe *PrefixExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(pe.Operator)
	out.WriteString(pe.Right.String())
	out.WriteString(")")

	return out.String()
}

type InfixExpression struct {
	Token    token.Token // the infix token i.e. +, -
	Operator string
	Right    Expression
	Left     Expression
}

func (ie *InfixExpression) expressionNode()      {}
func (ie *InfixExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *InfixExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString(" " + ie.Operator + " ")
	out.WriteString(ie.Right.String())
	out.WriteString(")")

	return out.String()
}

type Program struct {
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	} else {
		return ""
	}
}

func (p *Program) String() string {
	var out bytes.Buffer
	for _, s := range p.Statements {
		out.WriteString(s.String())
	}
	return out.String()
}

type IfExpression struct {
	Token       token.Token
	Condition   Expression
	Consequence *BlockStatement
	Alternative *BlockStatement
}

func (ie *IfExpression) expressionNode() {}
func (ie *IfExpression) TokenLiteral() string {
	return ie.Token.Literal
}

func (ie *IfExpression) String() string {
	var out bytes.Buffer

	out.WriteString("if ")
	out.WriteString(ie.Condition.String())
	out.WriteString("	{")
	out.WriteString(ie.Consequence.String())
	out.WriteString("	}")

	if ie.Alternative != nil {
		out.WriteString("else {")
		out.WriteString(ie.Alternative.String())
		out.WriteString("	}")
	}
	return out.String()
}

type BlockStatement struct {
	Token      token.Token // the { token
	Statements []Statement
}

func (bs *BlockStatement) TokenLiteral() string {
	return bs.Token.Literal
}

func (bs *BlockStatement) String() string {
	var out bytes.Buffer
	for _, s := range bs.Statements {
		out.WriteString(s.String())
	}
	return out.String()
}

type ForLoop struct {
	Token         token.Token // the for token
	InitStatement Statement
	Condition     Expression
	PostStatement Statement
	ForBody       *BlockStatement
}

func (fl *ForLoop) expressionNode() {}
func (fl *ForLoop) TokenLiteral() string {
	return fl.Token.Literal
}

func (fl *ForLoop) String() string {
	var out bytes.Buffer

	out.WriteString(fmt.Sprintf("%s (%s; %s; %s) {\n", fl.TokenLiteral(), fl.InitStatement.String(), fl.Condition.String(), fl.PostStatement.String()))
	for _, bodyStmt := range fl.ForBody.Statements {
		out.WriteString(fmt.Sprintf("\t%s\n", bodyStmt.String()))
	}
	out.WriteString("}")

	return out.String()
}

type FunctionLiteral struct {
	Token        token.Token // the fn token
	Parameters   []*Identifier
	FunctionBody *BlockStatement
}

func (fl *FunctionLiteral) expressionNode() {}
func (fl *FunctionLiteral) TokenLiteral() string {
	return fl.Token.Literal
}

func (fl *FunctionLiteral) String() string {
	var out bytes.Buffer

	out.WriteString(fl.TokenLiteral())
	out.WriteString("(")

	for i, p := range fl.Parameters {
		out.WriteString(p.String())
		if i != len(fl.Parameters)-1 {
			out.WriteString(", ")
		}
	}

	out.WriteString(") {")
	out.WriteString(fl.FunctionBody.String())
	out.WriteString("}")
	return out.String()
}

type CallExpression struct {
	Token token.Token // LPAREN
	// actually this can only be an Identifier which maps to an expression, or it can be a FunctionLiteral, but not any arbitrary expression
	Function  Expression
	Arguments []Expression
}

func (ce *CallExpression) expressionNode() {}
func (ce *CallExpression) TokenLiteral() string {
	return ce.Token.Literal
}

func (ce *CallExpression) String() string {
	var out bytes.Buffer

	out.WriteString(ce.Function.String())
	out.WriteString("(")
	for i, arg := range ce.Arguments {
		out.WriteString(arg.String())
		if i != len(ce.Arguments)-1 {
			out.WriteString(", ")
		}
	}
	out.WriteString(")")
	return out.String()
}

type IndexExpression struct {
	Token token.Token // LBRACKET
	Left  Expression  // this is the ident of the array, hash, or literal
	Index Expression
}

func (aie *IndexExpression) expressionNode() {}
func (aie *IndexExpression) TokenLiteral() string {
	return aie.Token.Literal
}

func (aie *IndexExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(aie.Left.String())
	out.WriteString("[")
	out.WriteString(aie.Index.String())
	out.WriteString("])")
	return out.String()
}

type HashLiteral struct {
	Token token.Token // { token
	Pairs map[Expression]Expression
}

func (h *HashLiteral) expressionNode() {}
func (h *HashLiteral) TokenLiteral() string {
	return h.Token.Literal
}

func (h *HashLiteral) String() string {
	var out bytes.Buffer

	out.WriteString("{")
	for k, v := range h.Pairs {
		out.WriteString(k.String())
		out.WriteString(": ")
		out.WriteString(v.String())
	}
	out.WriteString("}")
	return out.String()
}
