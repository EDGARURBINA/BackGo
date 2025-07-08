package interfaces

import "mongo-analyzer/domain/entities"

type Lexer interface {
	Tokenize(input string) ([]*entities.Token, error)
}
