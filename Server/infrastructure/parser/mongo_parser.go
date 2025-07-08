package parser

import (
	"fmt"
	"strconv"
	"strings"

	"mongo-analyzer/domain/entities"
)

type MongoParser struct {
	tokens   []*entities.Token
	position int
	current  *entities.Token
}

func NewMongoParser() *MongoParser {
	return &MongoParser{}
}

func (p *MongoParser) Parse(tokens []*entities.Token) (*entities.MongoCommand, error) {
	p.tokens = tokens
	p.position = 0
	p.current = p.tokens[0]

	command, err := p.parseCommand()
	if err != nil {
		return nil, err
	}

	command.TokenCount = len(tokens) - 1 // Excluir EOF
	return command, nil
}

func (p *MongoParser) parseCommand() (*entities.MongoCommand, error) {
	if p.current.Type == entities.USE {
		return p.parseUseCommand()
	}

	if p.current.Type == entities.DB {
		return p.parseDbCommand()
	}

	return nil, fmt.Errorf("comando no reconocido: %s", p.current.Value)
}

func (p *MongoParser) parseUseCommand() (*entities.MongoCommand, error) {
	p.advance() // skip 'use'

	if p.current.Type != entities.IDENTIFIER {
		return &entities.MongoCommand{
			IsValid: false,
			Errors:  []string{"Se esperaba nombre de base de datos después de 'use'"},
		}, nil
	}

	dbName := p.current.Value
	p.advance()

	// Verificar si hay dropDatabase después
	if p.current.Type == entities.DB {
		p.advance() // skip 'db'
		if p.current.Type == entities.DOT {
			p.advance() // skip '.'
			if p.current.Type == entities.FUNCTION && p.current.Value == "dropDatabase" {
				return &entities.MongoCommand{
					Type:     entities.DROP_DATABASE,
					Database: dbName,
					IsValid:  true,
				}, nil
			}
		}
	}

	return &entities.MongoCommand{
		Type:     entities.USE_DATABASE,
		Database: dbName,
		IsValid:  true,
	}, nil
}

func (p *MongoParser) parseDbCommand() (*entities.MongoCommand, error) {
	p.advance() // skip 'db'

	if p.current.Type != entities.DOT {
		return &entities.MongoCommand{
			IsValid: false,
			Errors:  []string{"Se esperaba '.' después de 'db'"},
		}, nil
	}
	p.advance() // skip '.'

	// ✅ NUEVO: Manejar db.createCollection() y db.dropDatabase()
	if p.current.Type == entities.FUNCTION {
		switch p.current.Value {
		case "createCollection":
			return p.parseCreateCollection()
		case "dropDatabase":
			return p.parseDropDatabase() // ✅ AGREGADO
		default:
			return &entities.MongoCommand{
				IsValid: false,
				Errors:  []string{fmt.Sprintf("Función no reconocida: %s", p.current.Value)},
			}, nil
		}
	}

	if p.current.Type == entities.IDENTIFIER {
		collection := p.current.Value
		p.advance()

		if p.current.Type != entities.DOT {
			return &entities.MongoCommand{
				IsValid: false,
				Errors:  []string{"Se esperaba '.' después del nombre de la colección"},
			}, nil
		}
		p.advance() // skip '.'

		if p.current.Type != entities.FUNCTION {
			return &entities.MongoCommand{
				IsValid: false,
				Errors:  []string{"Se esperaba función después de '.'"},
			}, nil
		}

		switch p.current.Value {
		case "insertOne":
			return p.parseInsertOne(collection)
		case "find":
			return p.parseFind(collection)
		case "updateOne":
			return p.parseUpdateOne(collection)
		case "deleteOne":
			return p.parseDeleteOne(collection)
		case "drop":
			return p.parseDrop(collection)
		default:
			return &entities.MongoCommand{
				IsValid: false,
				Errors:  []string{fmt.Sprintf("Función no reconocida: %s", p.current.Value)},
			}, nil
		}
	}

	return &entities.MongoCommand{
		IsValid: false,
		Errors:  []string{"Comando db inválido"},
	}, nil
}

func (p *MongoParser) parseCreateCollection() (*entities.MongoCommand, error) {
	p.advance() // skip 'createCollection'

	if p.current.Type != entities.LEFT_PAREN {
		return &entities.MongoCommand{
			IsValid: false,
			Errors:  []string{"Se esperaba '(' después de createCollection"},
		}, nil
	}
	p.advance()

	if p.current.Type != entities.STRING {
		return &entities.MongoCommand{
			IsValid: false,
			Errors:  []string{"Se esperaba nombre de colección como string"},
		}, nil
	}

	collection := p.current.Value
	p.advance()

	if p.current.Type != entities.RIGHT_PAREN {
		return &entities.MongoCommand{
			IsValid: false,
			Errors:  []string{"Se esperaba ')' después del nombre de la colección"},
		}, nil
	}

	return &entities.MongoCommand{
		Type:       entities.CREATE_COLLECTION,
		Collection: collection,
		IsValid:    true,
	}, nil
}

func (p *MongoParser) parseInsertOne(collection string) (*entities.MongoCommand, error) {
	p.advance() // skip 'insertOne'

	if p.current.Type != entities.LEFT_PAREN {
		return &entities.MongoCommand{
			IsValid: false,
			Errors:  []string{"Se esperaba '(' después de insertOne"},
		}, nil
	}
	p.advance()

	document, err := p.parseDocument()
	if err != nil {
		return &entities.MongoCommand{
			IsValid: false,
			Errors:  []string{err.Error()},
		}, nil
	}

	if p.current.Type != entities.RIGHT_PAREN {
		return &entities.MongoCommand{
			IsValid: false,
			Errors:  []string{"Se esperaba ')' después del documento"},
		}, nil
	}

	return &entities.MongoCommand{
		Type:       entities.INSERT_ONE,
		Collection: collection,
		Document:   document,
		IsValid:    true,
	}, nil
}

func (p *MongoParser) parseFind(collection string) (*entities.MongoCommand, error) {
	p.advance() // skip 'find'

	if p.current.Type != entities.LEFT_PAREN {
		return &entities.MongoCommand{
			IsValid: false,
			Errors:  []string{"Se esperaba '(' después de find"},
		}, nil
	}
	p.advance()

	var filter map[string]interface{}
	var err error

	// find() puede tener filtro opcional
	if p.current.Type == entities.LEFT_BRACE {
		filter, err = p.parseDocument()
		if err != nil {
			return &entities.MongoCommand{
				IsValid: false,
				Errors:  []string{err.Error()},
			}, nil
		}
	}

	if p.current.Type != entities.RIGHT_PAREN {
		return &entities.MongoCommand{
			IsValid: false,
			Errors:  []string{"Se esperaba ')' después de find"},
		}, nil
	}

	return &entities.MongoCommand{
		Type:       entities.FIND,
		Collection: collection,
		Filter:     filter,
		IsValid:    true,
	}, nil
}

func (p *MongoParser) parseUpdateOne(collection string) (*entities.MongoCommand, error) {
	p.advance() // skip 'updateOne'

	if p.current.Type != entities.LEFT_PAREN {
		return &entities.MongoCommand{
			IsValid: false,
			Errors:  []string{"Se esperaba '(' después de updateOne"},
		}, nil
	}
	p.advance()

	// Parse filter
	filter, err := p.parseDocument()
	if err != nil {
		return &entities.MongoCommand{
			IsValid: false,
			Errors:  []string{"Error en filtro: " + err.Error()},
		}, nil
	}

	if p.current.Type != entities.COMMA {
		return &entities.MongoCommand{
			IsValid: false,
			Errors:  []string{"Se esperaba ',' entre filtro y actualización"},
		}, nil
	}
	p.advance()

	// Parse update
	update, err := p.parseDocument()
	if err != nil {
		return &entities.MongoCommand{
			IsValid: false,
			Errors:  []string{"Error en actualización: " + err.Error()},
		}, nil
	}

	if p.current.Type != entities.RIGHT_PAREN {
		return &entities.MongoCommand{
			IsValid: false,
			Errors:  []string{"Se esperaba ')' después de updateOne"},
		}, nil
	}

	return &entities.MongoCommand{
		Type:       entities.UPDATE_ONE,
		Collection: collection,
		Filter:     filter,
		Update:     update,
		IsValid:    true,
	}, nil
}

func (p *MongoParser) parseDeleteOne(collection string) (*entities.MongoCommand, error) {
	p.advance() // skip 'deleteOne'

	if p.current.Type != entities.LEFT_PAREN {
		return &entities.MongoCommand{
			IsValid: false,
			Errors:  []string{"Se esperaba '(' después de deleteOne"},
		}, nil
	}
	p.advance()

	filter, err := p.parseDocument()
	if err != nil {
		return &entities.MongoCommand{
			IsValid: false,
			Errors:  []string{err.Error()},
		}, nil
	}

	if p.current.Type != entities.RIGHT_PAREN {
		return &entities.MongoCommand{
			IsValid: false,
			Errors:  []string{"Se esperaba ')' después del filtro"},
		}, nil
	}

	return &entities.MongoCommand{
		Type:       entities.DELETE_ONE,
		Collection: collection,
		Filter:     filter,
		IsValid:    true,
	}, nil
}

func (p *MongoParser) parseDrop(collection string) (*entities.MongoCommand, error) {
	p.advance() // skip 'drop'

	if p.current.Type != entities.LEFT_PAREN {
		return &entities.MongoCommand{
			IsValid: false,
			Errors:  []string{"Se esperaba '(' después de drop"},
		}, nil
	}
	p.advance()

	if p.current.Type != entities.RIGHT_PAREN {
		return &entities.MongoCommand{
			IsValid: false,
			Errors:  []string{"Se esperaba ')' después de drop"},
		}, nil
	}

	return &entities.MongoCommand{
		Type:       entities.DROP_COLLECTION,
		Collection: collection,
		IsValid:    true,
	}, nil
}

// ✅ NUEVA FUNCIÓN: parseDropDatabase para db.dropDatabase()
func (p *MongoParser) parseDropDatabase() (*entities.MongoCommand, error) {
	p.advance() // skip 'dropDatabase'

	if p.current.Type != entities.LEFT_PAREN {
		return &entities.MongoCommand{
			IsValid: false,
			Errors:  []string{"Se esperaba '(' después de dropDatabase"},
		}, nil
	}
	p.advance()

	if p.current.Type != entities.RIGHT_PAREN {
		return &entities.MongoCommand{
			IsValid: false,
			Errors:  []string{"Se esperaba ')' después de dropDatabase"},
		}, nil
	}
	p.advance()

	return &entities.MongoCommand{
		Type:    entities.DROP_DATABASE,
		IsValid: true,
		// Nota: Database se establecerá en el executor basándose en la DB actual
	}, nil
}

func (p *MongoParser) parseDocument() (map[string]interface{}, error) {
	if p.current.Type != entities.LEFT_BRACE {
		return nil, fmt.Errorf("se esperaba '{' al inicio del documento")
	}
	p.advance()

	document := make(map[string]interface{})

	// Documento vacío
	if p.current.Type == entities.RIGHT_BRACE {
		p.advance()
		return document, nil
	}

	for {
		// Parse key - ✅ MEJORADO: Aceptar tanto strings como identificadores y operadores $
		var key string
		
		if p.current.Type == entities.STRING {
			key = p.current.Value
			p.advance()
		} else if p.current.Type == entities.IDENTIFIER {
			key = p.current.Value
			p.advance()
		} else if p.current.Type == entities.DOLLAR_SIGN {
			// ✅ NUEVO: Manejar operadores como $set, $inc, etc.
			p.advance() // skip '$'
			if p.current.Type != entities.IDENTIFIER {
				return nil, fmt.Errorf("se esperaba identificador después de '$'")
			}
			key = "$" + p.current.Value
			p.advance()
		} else {
			return nil, fmt.Errorf("se esperaba string, identificador o operador $ como clave")
		}

		if p.current.Type != entities.COLON {
			return nil, fmt.Errorf("se esperaba ':' después de la clave")
		}
		p.advance()

		// Parse value
		value, err := p.parseValue()
		if err != nil {
			return nil, err
		}

		document[key] = value

		if p.current.Type == entities.RIGHT_BRACE {
			p.advance()
			break
		}

		if p.current.Type != entities.COMMA {
			return nil, fmt.Errorf("se esperaba ',' o '}' en el documento")
		}
		p.advance()
	}

	return document, nil
}

// ✅ MEJORADO: parseValue para manejar mejor los tipos
func (p *MongoParser) parseValue() (interface{}, error) {
	switch p.current.Type {
	case entities.STRING:
		value := p.current.Value
		p.advance()
		return value, nil
	case entities.NUMBER:
		value := p.current.Value
		p.advance()
		if strings.Contains(value, ".") {
			return strconv.ParseFloat(value, 64)
		}
		return strconv.Atoi(value)
	case entities.LEFT_BRACE:
		return p.parseDocument()
	case entities.IDENTIFIER:
		// ✅ NUEVO: Permitir identificadores como valores (para campos sin comillas)
		value := p.current.Value
		p.advance()
		return value, nil
	case entities.DOLLAR_SIGN:
		// ✅ MEJORADO: Manejar operadores $ como valores
		p.advance()
		if p.current.Type != entities.IDENTIFIER {
			return nil, fmt.Errorf("se esperaba identificador después de '$'")
		}
		operator := "$" + p.current.Value
		p.advance()
		return operator, nil
	default:
		return nil, fmt.Errorf("valor no válido: %s", p.current.Value)
	}
}

func (p *MongoParser) advance() {
	if p.position < len(p.tokens)-1 {
		p.position++
		p.current = p.tokens[p.position]
	}
}