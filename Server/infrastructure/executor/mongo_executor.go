package executor

import (
	"context"
	"fmt"
	"time"

	"mongo-analyzer/domain/entities"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoExecutor struct {
	client         *mongo.Client
	connectionURI  string
	currentDB      string
	timeout        time.Duration
}

func NewMongoExecutor(connectionURI string) *MongoExecutor {
	return &MongoExecutor{
		connectionURI: connectionURI,
		timeout:       10 * time.Second,
	}
}

func (e *MongoExecutor) Connect() error {
	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(e.connectionURI))
	if err != nil {
		return fmt.Errorf("error conectando a MongoDB: %v", err)
	}

	
	if err := client.Ping(ctx, nil); err != nil {
		return fmt.Errorf("error haciendo ping a MongoDB: %v", err)
	}

	e.client = client
	return nil
}

func (e *MongoExecutor) Close() error {
	if e.client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
		defer cancel()
		return e.client.Disconnect(ctx)
	}
	return nil
}

func (e *MongoExecutor) Execute(command *entities.MongoCommand) (interface{}, error) {
	if e.client == nil {
		return nil, fmt.Errorf("no hay conexión a MongoDB")
	}

	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	defer cancel()

	switch command.Type {
	case entities.USE_DATABASE:
		return e.executeUseDatabase(ctx, command)
	case entities.CREATE_COLLECTION:
		return e.executeCreateCollection(ctx, command)
	case entities.INSERT_ONE:
		return e.executeInsertOne(ctx, command)
	case entities.FIND:
		return e.executeFind(ctx, command)
	case entities.UPDATE_ONE:
		return e.executeUpdateOne(ctx, command)
	case entities.DELETE_ONE:
		return e.executeDeleteOne(ctx, command)
	case entities.DROP_COLLECTION:
		return e.executeDropCollection(ctx, command)
	case entities.DROP_DATABASE:
		return e.executeDropDatabase(ctx, command)
	default:
		return nil, fmt.Errorf("tipo de comando no soportado")
	}
}

func (e *MongoExecutor) executeUseDatabase(ctx context.Context, command *entities.MongoCommand) (interface{}, error) {
	e.currentDB = command.Database
	
	// Verificar si la base de datos existe o crearla
	db := e.client.Database(command.Database)
	
	// Crear una colección temporal para forzar la creación de la DB
	tempCollection := db.Collection("__temp__")
	_, err := tempCollection.InsertOne(ctx, bson.M{"temp": true})
	if err != nil {
		return nil, err
	}
	
	// Eliminar la colección temporal
	tempCollection.Drop(ctx)
	
	return map[string]interface{}{
		"message":  fmt.Sprintf("Cambiado a base de datos '%s'", command.Database),
		"database": command.Database,
	}, nil
}

func (e *MongoExecutor) executeCreateCollection(ctx context.Context, command *entities.MongoCommand) (interface{}, error) {
	if e.currentDB == "" {
		return nil, fmt.Errorf("no hay base de datos seleccionada. Usa 'use nombreDB' primero")
	}

	db := e.client.Database(e.currentDB)
	err := db.CreateCollection(ctx, command.Collection)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"message":    fmt.Sprintf("Colección '%s' creada exitosamente", command.Collection),
		"collection": command.Collection,
		"database":   e.currentDB,
	}, nil
}

func (e *MongoExecutor) executeInsertOne(ctx context.Context, command *entities.MongoCommand) (interface{}, error) {
	if e.currentDB == "" {
		return nil, fmt.Errorf("no hay base de datos seleccionada. Usa 'use nombreDB' primero")
	}

	collection := e.client.Database(e.currentDB).Collection(command.Collection)
	result, err := collection.InsertOne(ctx, command.Document)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"message":     "Documento insertado exitosamente",
		"insertedId":  result.InsertedID,
		"collection":  command.Collection,
		"database":    e.currentDB,
	}, nil
}

func (e *MongoExecutor) executeFind(ctx context.Context, command *entities.MongoCommand) (interface{}, error) {
	if e.currentDB == "" {
		return nil, fmt.Errorf("no hay base de datos seleccionada. Usa 'use nombreDB' primero")
	}

	collection := e.client.Database(e.currentDB).Collection(command.Collection)
	
	filter := bson.M{}
	if command.Filter != nil {
		filter = command.Filter
	}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"message":    fmt.Sprintf("Encontrados %d documentos", len(results)),
		"documents":  results,
		"count":      len(results),
		"collection": command.Collection,
		"database":   e.currentDB,
	}, nil
}

func (e *MongoExecutor) executeUpdateOne(ctx context.Context, command *entities.MongoCommand) (interface{}, error) {
	if e.currentDB == "" {
		return nil, fmt.Errorf("no hay base de datos seleccionada. Usa 'use nombreDB' primero")
	}

	collection := e.client.Database(e.currentDB).Collection(command.Collection)
	result, err := collection.UpdateOne(ctx, command.Filter, command.Update)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"message":       "Actualización completada",
		"matchedCount":  result.MatchedCount,
		"modifiedCount": result.ModifiedCount,
		"collection":    command.Collection,
		"database":      e.currentDB,
	}, nil
}

func (e *MongoExecutor) executeDeleteOne(ctx context.Context, command *entities.MongoCommand) (interface{}, error) {
	if e.currentDB == "" {
		return nil, fmt.Errorf("no hay base de datos seleccionada. Usa 'use nombreDB' primero")
	}

	collection := e.client.Database(e.currentDB).Collection(command.Collection)
	result, err := collection.DeleteOne(ctx, command.Filter)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"message":      "Eliminación completada",
		"deletedCount": result.DeletedCount,
		"collection":   command.Collection,
		"database":     e.currentDB,
	}, nil
}

func (e *MongoExecutor) executeDropCollection(ctx context.Context, command *entities.MongoCommand) (interface{}, error) {
	if e.currentDB == "" {
		return nil, fmt.Errorf("no hay base de datos seleccionada. Usa 'use nombreDB' primero")
	}

	collection := e.client.Database(e.currentDB).Collection(command.Collection)
	err := collection.Drop(ctx)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"message":    fmt.Sprintf("Colección '%s' eliminada exitosamente", command.Collection),
		"collection": command.Collection,
		"database":   e.currentDB,
	}, nil
}

func (e *MongoExecutor) executeDropDatabase(ctx context.Context, command *entities.MongoCommand) (interface{}, error) {
	// ✅ MEJORADO: Si no hay database especificada, usar la actual
	var databaseName string
	if command.Database != "" {
		databaseName = command.Database
	} else if e.currentDB != "" {
		databaseName = e.currentDB
	} else {
		return nil, fmt.Errorf("no hay base de datos especificada o seleccionada. Usa 'use nombreDB' primero o especifica la base de datos")
	}

	db := e.client.Database(databaseName)
	err := db.Drop(ctx)
	if err != nil {
		return nil, err
	}

	// ✅ Si se eliminó la DB actual, limpiar currentDB
	if databaseName == e.currentDB {
		e.currentDB = ""
	}

	return map[string]interface{}{
		"message":  fmt.Sprintf("Base de datos '%s' eliminada exitosamente", databaseName),
		"database": databaseName,
	}, nil
}