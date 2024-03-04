package token

type TokenType string

type Token struct {
	Type    TokenType
	Literal string
}

const (
	// ILLEGAL represents a token type for errors or unknown tokens.
	// iota is used to automatically increment the numeric value of the constants
	// that follow, providing a unique value for each TokenType.
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	// Identifiers + literals
	// Identifiers are user-defined names for variables, functions, etc.
	// Literals are fixed values like numbers, strings, etc.
	IDENT  = "IDENT"
	INT    = "INT"
	STRING = "STRING"

	// Operators
	// Operators are special symbols that represent computations like addition, subtraction, etc.
	// The operator tokens are the actual characters like +, -, etc.
	PLUS     = "+"
	ASSIGN   = "="
	MINUS    = "-"
	BANG     = "!"
	ASTERISK = "*"
	SLASH    = "/"

	LT     = "<"
	GT     = ">"
	EQ     = "=="
	NOT_EQ = "!="

	// Delimiters
	// Delimiters are special symbols that are used to separate tokens.
	// The delimiter tokens are the actual characters like (, ), etc.
	COMMA     = ","
	SEMICOLON = ";"
	LPAREN    = "("
	RPAREN    = ")"
	LBRACE    = "{"
	RBRACE    = "}"

	// Keywords
	// Keywords are reserved words that have special meaning in the language.
	// The keyword tokens are the actual words like fn, let, etc.
	FUNCTION = "FUNCTION"
	LET      = "LET"
	IF       = "IF"
	ELSE     = "ELSE"
	RETURN   = "RETURN"
	TRUE     = "TRUE"
	FALSE    = "FALSE"
)

var keywords = map[string]TokenType{
	"fn":     FUNCTION,
	"let":    LET,
	"if":     IF,
	"else":   ELSE,
	"true":   TRUE,
	"false":  FALSE,
	"return": RETURN,
}

func LookupIdent(ident string) TokenType {
	if toktype, ok := keywords[ident]; ok {
		return toktype
	}
	return IDENT
}
