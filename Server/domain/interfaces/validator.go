package interfaces

import "mongo-analyzer/domain/entities"

type Validator interface {
	ValidateSemantics(command *entities.MongoCommand) error
}
