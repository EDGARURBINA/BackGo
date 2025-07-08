package lexer

import (
	"fmt"
	"strings"
	"unicode"

	"mongo-analyzer/domain/entities"
)

type MongoLexer struct {
	input    string
	position int
	line     int
	column   int
}

func NewMongoLexer() *MongoLexer {
	return &MongoLexer{
		line:   1,
		column: 1,
	}
}

func (l *MongoLexer) Tokenize(input string) ([]*entities.Token, error) {
	l.input = strings.TrimSpace(input)
	l.position = 0
	l.line = 1
	l.column = 1

	var tokens []*entities.Token

	for l.position < len(l.input) {
		token := l.nextToken()
		if token.Type == entities.INVALID {
			return nil, fmt.Errorf("token inválido en posición %d: '%s'", token.Position, token.Value)
		}
		if token.Type != entities.EOF {
			tokens = append(tokens, token)
		}
	}

	// Agregar EOF
	tokens = append(tokens, &entities.Token{
		Type:     entities.EOF,
		Position: l.position,
		Line:     l.line,
		Column:   l.column,
	})

	return tokens, nil
}

func (l *MongoLexer) nextToken() *entities.Token {
	l.skipWhitespace()

	if l.position >= len(l.input) {
		return &entities.Token{Type: entities.EOF, Position: l.position}
	}

	start := l.position
	ch := l.input[l.position]

	switch ch {
	case '.':
		l.advance()
		return &entities.Token{Type: entities.DOT, Value: ".", Position: start, Line: l.line, Column: l.column - 1}
	case '(':
		l.advance()
		return &entities.Token{Type: entities.LEFT_PAREN, Value: "(", Position: start, Line: l.line, Column: l.column - 1}
	case ')':
		l.advance()
		return &entities.Token{Type: entities.RIGHT_PAREN, Value: ")", Position: start, Line: l.line, Column: l.column - 1}
	case '{':
		l.advance()
		return &entities.Token{Type: entities.LEFT_BRACE, Value: "{", Position: start, Line: l.line, Column: l.column - 1}
	case '}':
		l.advance()
		return &entities.Token{Type: entities.RIGHT_BRACE, Value: "}", Position: start, Line: l.line, Column: l.column - 1}
	case ',':
		l.advance()
		return &entities.Token{Type: entities.COMMA, Value: ",", Position: start, Line: l.line, Column: l.column - 1}
	case ':':
		l.advance()
		return &entities.Token{Type: entities.COLON, Value: ":", Position: start, Line: l.line, Column: l.column - 1}
	case '$':
		l.advance()
		return &entities.Token{Type: entities.DOLLAR_SIGN, Value: "$", Position: start, Line: l.line, Column: l.column - 1}
	case '"':
		return l.readString()
	}

	if unicode.IsDigit(rune(ch)) {
		return l.readNumber()
	}

	if unicode.IsLetter(rune(ch)) || ch == '_' {
		return l.readIdentifier()
	}

	l.advance()
	return &entities.Token{Type: entities.INVALID, Value: string(ch), Position: start}
}

func (l *MongoLexer) readString() *entities.Token {
	start := l.position
	l.advance() // skip opening quote

	value := ""
	for l.position < len(l.input) && l.input[l.position] != '"' {
		value += string(l.input[l.position])
		l.advance()
	}

	if l.position >= len(l.input) {
		return &entities.Token{Type: entities.INVALID, Value: value, Position: start}
	}

	l.advance() // skip closing quote
	return &entities.Token{Type: entities.STRING, Value: value, Position: start, Line: l.line, Column: l.column - len(value) - 2}
}

func (l *MongoLexer) readNumber() *entities.Token {
	start := l.position
	value := ""

	for l.position < len(l.input) && (unicode.IsDigit(rune(l.input[l.position])) || l.input[l.position] == '.') {
		value += string(l.input[l.position])
		l.advance()
	}

	return &entities.Token{Type: entities.NUMBER, Value: value, Position: start, Line: l.line, Column: l.column - len(value)}
}

func (l *MongoLexer) readIdentifier() *entities.Token {
	start := l.position
	value := ""

	for l.position < len(l.input) && (unicode.IsLetter(rune(l.input[l.position])) || unicode.IsDigit(rune(l.input[l.position])) || l.input[l.position] == '_') {
		value += string(l.input[l.position])
		l.advance()
	}

	tokenType := l.getIdentifierType(value)
	return &entities.Token{Type: tokenType, Value: value, Position: start, Line: l.line, Column: l.column - len(value)}
}

func (l *MongoLexer) getIdentifierType(value string) entities.TokenType {
	switch value {
	case "use":
		return entities.USE
	case "db":
		return entities.DB
	default:
		// Verificar si es una función conocida
		functions := []string{"createCollection", "insertOne", "find", "updateOne", "deleteOne", "drop", "dropDatabase"}
		for _, fn := range functions {
			if value == fn {
				return entities.FUNCTION
			}
		}
		return entities.IDENTIFIER
	}
}

func (l *MongoLexer) skipWhitespace() {
	for l.position < len(l.input) && unicode.IsSpace(rune(l.input[l.position])) {
		if l.input[l.position] == '\n' {
			l.line++
			l.column = 1
		} else {
			l.column++
		}
		l.position++
	}
}

func (l *MongoLexer) advance() {
	if l.position < len(l.input) {
		l.position++
		l.column++
	}
}