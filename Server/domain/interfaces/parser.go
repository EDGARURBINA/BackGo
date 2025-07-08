package interfaces

import "mongo-analyzer/domain/entities"

type Parser interface {
	Parse(tokens []*entities.Token) (*entities.MongoCommand, error)
}
