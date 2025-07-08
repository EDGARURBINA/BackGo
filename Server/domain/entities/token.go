package entities

type TokenType int

const (
	USE TokenType = iota
	DB
	DOT
	IDENTIFIER
	FUNCTION
	LEFT_PAREN
	RIGHT_PAREN
	LEFT_BRACE
	RIGHT_BRACE
	STRING
	NUMBER
	COMMA
	COLON
	DOLLAR_SIGN
	EOF
	INVALID
)

type Token struct {
	Type     TokenType
	Value    string
	Position int
	Line     int
	Column   int
}
