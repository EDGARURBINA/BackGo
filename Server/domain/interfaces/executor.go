package interfaces

import "mongo-analyzer/domain/entities"

type MongoExecutor interface {
	Execute(command *entities.MongoCommand) (interface{}, error)
	Connect() error
	Close() error
}