package validator

import (
	"fmt"
	"regexp"
	"mongo-analyzer/domain/entities"
)

type MongoValidator struct{}

func NewMongoValidator() *MongoValidator {
	return &MongoValidator{}
}

func (v *MongoValidator) ValidateSemantics(command *entities.MongoCommand) error {
	if !command.IsValid {
		return fmt.Errorf("comando sintácticamente inválido")
	}

	switch command.Type {
	case entities.USE_DATABASE:
		return v.validateDatabaseName(command.Database)
	case entities.CREATE_COLLECTION:
		return v.validateCollectionName(command.Collection)
	case entities.INSERT_ONE:
		return v.validateInsertDocument(command.Document)
	case entities.UPDATE_ONE:
		return v.validateUpdateCommand(command.Filter, command.Update)
	case entities.DELETE_ONE:
		return v.validateDeleteCommand(command.Filter)
	}

	return nil
}

func (v *MongoValidator) validateDatabaseName(name string) error {
	if name == "" {
		return fmt.Errorf("el nombre de la base de datos no puede estar vacío")
	}

	// MongoDB database name restrictions
	invalidChars := regexp.MustCompile(`[/\\. "$<>:|?*]`)
	if invalidChars.MatchString(name) {
		return fmt.Errorf("el nombre de la base de datos contiene caracteres inválidos")
	}

	if len(name) > 64 {
		return fmt.Errorf("el nombre de la base de datos es demasiado largo (máximo 64 caracteres)")
	}

	return nil
}

func (v *MongoValidator) validateCollectionName(name string) error {
	if name == "" {
		return fmt.Errorf("el nombre de la colección no puede estar vacío")
	}

	if name[0] == '$' {
		return fmt.Errorf("el nombre de la colección no puede comenzar con '$'")
	}

	return nil
}

func (v *MongoValidator) validateInsertDocument(doc map[string]interface{}) error {
	if len(doc) == 0 {
		return fmt.Errorf("el documento a insertar no puede estar vacío")
	}

	// Validar que las claves no contengan caracteres especiales
	for key := range doc {
		if key == "" {
			return fmt.Errorf("las claves del documento no pueden estar vacías")
		}
		if key[0] == '$' {
			return fmt.Errorf("las claves del documento no pueden comenzar con '$'")
		}
	}

	return nil
}

func (v *MongoValidator) validateUpdateCommand(filter, update map[string]interface{}) error {
	if len(filter) == 0 {
		return fmt.Errorf("el filtro de actualización no puede estar vacío")
	}

	if len(update) == 0 {
		return fmt.Errorf("la actualización no puede estar vacía")
	}

	// Validar que update tenga operadores válidos
	hasValidOperator := false
	validOperators := []string{"$set", "$unset", "$inc", "$push", "$pull"}
	
	for key := range update {
		for _, op := range validOperators {
			if key == op {
				hasValidOperator = true
				break
			}
		}
	}

	if !hasValidOperator {
		return fmt.Errorf("la actualización debe contener al menos un operador válido ($set, $unset, $inc, etc.)")
	}

	return nil
}

func (v *MongoValidator) validateDeleteCommand(filter map[string]interface{}) error {
	if len(filter) == 0 {
		return fmt.Errorf("el filtro de eliminación no puede estar vacío")
	}

	return nil
}
