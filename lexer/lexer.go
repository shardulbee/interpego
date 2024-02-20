package lexer

import (
	"interpego/token"
)

type Lexer struct {
	input   string
	curpos  int  // index of the char we are currently reading
	ch      byte // char we are currently reading
	nextpos int  // index of the next char to read
}

// gracefully handles reading end of input
func (l *Lexer) readChar() {
	if l.nextpos >= len(l.input) {
		// we define 0 to be an "EOF" char
		l.ch = 0
	} else {
		l.ch = l.input[l.nextpos]
	}
	l.curpos = l.nextpos
	l.nextpos += 1
}

func (l *Lexer) peekChar() byte {
	if l.nextpos >= len(l.input) {
		return 0
	} else {
		return l.input[l.nextpos]
	}
}

func New(input string) *Lexer {
	l := &Lexer{input: input}
	l.readChar()
	return l
}

func (l *Lexer) NextToken() token.Token {
	var tok token.Token
	l.skipWhitespace()
	switch l.ch {
	case '=':
		if l.peekChar() == '=' {
			curchar := l.ch
			l.readChar()
			nextchar := l.ch
			lit := string(curchar) + string(nextchar)
			tok = token.Token{Type: token.EQ, Literal: lit}
		} else {
			tok = token.Token{Type: token.ASSIGN, Literal: string(l.ch)}
		}
	case '+':
		tok = token.Token{Type: token.PLUS, Literal: string(l.ch)}
	case ';':
		tok = token.Token{Type: token.SEMICOLON, Literal: string(l.ch)}
	case '(':
		tok = token.Token{Type: token.LPAREN, Literal: string(l.ch)}
	case ')':
		tok = token.Token{Type: token.RPAREN, Literal: string(l.ch)}
	case '{':
		tok = token.Token{Type: token.LBRACE, Literal: string(l.ch)}
	case '}':
		tok = token.Token{Type: token.RBRACE, Literal: string(l.ch)}
	case ',':
		tok = token.Token{Type: token.COMMA, Literal: string(l.ch)}
	case '-':
		tok = token.Token{Type: token.MINUS, Literal: string(l.ch)}
	case '!':
		if l.peekChar() == '=' {
			curchar := l.ch
			l.readChar()
			nextchar := l.ch
			lit := string(curchar) + string(nextchar)
			tok = token.Token{Type: token.NOT_EQ, Literal: lit}
		} else {
			tok = token.Token{Type: token.BANG, Literal: string(l.ch)}
		}
	case '*':
		tok = token.Token{Type: token.ASTERISK, Literal: string(l.ch)}
	case '/':
		tok = token.Token{Type: token.SLASH, Literal: string(l.ch)}
	case '<':
		tok = token.Token{Type: token.LT, Literal: string(l.ch)}
	case '>':
		tok = token.Token{Type: token.GT, Literal: string(l.ch)}
	case 0:
		tok = token.Token{Type: token.EOF, Literal: ""}
	default:
		if isLetter(l.ch) {
			ident := l.readIdentifier()

			tok.Literal = ident
			tok.Type = token.LookupIdent(tok.Literal)
			return tok
		} else if isDigit(l.ch) {
			tok.Type = token.INT
			tok.Literal = l.readNumber()
			return tok
		} else {
			tok = token.Token{Type: token.ILLEGAL, Literal: ""}
		}
	}
	l.readChar()
	return tok
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func (l *Lexer) readNumber() string {
	startpos := l.curpos
	for isDigit(l.ch) {
		l.readChar()
	}

	return l.input[startpos:l.curpos]
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

func (l *Lexer) readIdentifier() string {
	startpos := l.curpos
	for isLetter(l.ch) {
		l.readChar()
	}

	return l.input[startpos:l.curpos]
}

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}
