package services

import (
	"fmt"
	"strings"

	"mongo-analyzer/domain/entities"
	"mongo-analyzer/domain/interfaces"
)

type MongoAnalyzerService struct {
	lexer     interfaces.Lexer
	parser    interfaces.Parser
	validator interfaces.Validator
	executor  interfaces.MongoExecutor
}

func NewMongoAnalyzerService(
	lexer interfaces.Lexer,
	parser interfaces.Parser,
	validator interfaces.Validator,
	executor interfaces.MongoExecutor,
) *MongoAnalyzerService {
	return &MongoAnalyzerService{
		lexer:     lexer,
		parser:    parser,
		validator: validator,
		executor:  executor,
	}
}

func (s *MongoAnalyzerService) Analyze(input string) (*entities.AnalysisResult, error) {
	// Fase 1: Análisis Léxico
	tokens, err := s.lexer.Tokenize(input)
	if err != nil {
		return &entities.AnalysisResult{
			IsValid:      false,
			Errors:       []string{"Error léxico: " + err.Error()},
			SuggestedFix: s.generateLexicalFix(input, err),
		}, nil
	}

	// Fase 2: Análisis Sintáctico
	command, err := s.parser.Parse(tokens)
	if err != nil {
		return &entities.AnalysisResult{
			IsValid:      false,
			Errors:       []string{"Error sintáctico: " + err.Error()},
			TokenCount:   len(tokens) - 1, // Excluir EOF
			SuggestedFix: s.generateSyntacticFix(input, err),
		}, nil
	}

	if !command.IsValid {
		return &entities.AnalysisResult{
			Command:      command,
			IsValid:      false,
			Errors:       command.Errors,
			TokenCount:   command.TokenCount,
			SuggestedFix: s.generateSyntacticFixFromCommand(command),
		}, nil
	}

	// Fase 3: Análisis Semántico
	if err := s.validator.ValidateSemantics(command); err != nil {
		return &entities.AnalysisResult{
			Command:      command,
			IsValid:      false,
			Errors:       []string{"Error semántico: " + err.Error()},
			TokenCount:   command.TokenCount,
			SuggestedFix: s.generateSemanticFix(command, err),
		}, nil
	}

	// Fase 4: Ejecución
	var executionResult interface{}
	var executionError error

	if s.executor != nil {
		executionResult, executionError = s.executor.Execute(command)
	}

	return &entities.AnalysisResult{
		Command:         command,
		IsValid:         true,
		TokenCount:      command.TokenCount,
		ExecutionResult: executionResult,
		ExecutionError:  executionError,
	}, nil
}

// ✅ CORREGIDO: Usar _ para parámetros intencionalmente no usados
func (s *MongoAnalyzerService) generateLexicalFix(_ string, err error) string {
	// Sugerencias básicas para errores léxicos
	if strings.Contains(err.Error(), "token inválido") {
		return "Verifica caracteres especiales. Ejemplo correcto: db.usuarios.find()"
	}
	return "Revisa la sintaxis del comando"
}

// ✅ CORREGIDO: Usar _ para parámetros intencionalmente no usados
func (s *MongoAnalyzerService) generateSyntacticFix(_ string, err error) string {
	errorMsg := err.Error()
	
	if strings.Contains(errorMsg, "Se esperaba '.'") {
		return "Agrega un punto después de 'db': db.nombreColeccion.funcion()"
	}
	if strings.Contains(errorMsg, "Se esperaba '('") {
		return "Agrega paréntesis después de la función: funcion()"
	}
	if strings.Contains(errorMsg, "Se esperaba ')'") {
		return "Cierra los paréntesis: funcion(...)"
	}
	if strings.Contains(errorMsg, "Se esperaba '{'") {
		return "Usa llaves para objetos: { campo: valor }"
	}
	
	return "Revisa la sintaxis del comando MongoDB"
}

// ✅ CORREGIDO: Arreglar fmt.Errorf con format string no constante
func (s *MongoAnalyzerService) generateSyntacticFixFromCommand(command *entities.MongoCommand) string {
	if len(command.Errors) > 0 {
		// Opción 1: Crear error con format string correcto
		return s.generateSyntacticFix("", fmt.Errorf("%s", command.Errors[0]))
		
		// Opción 2: Pasar directamente el string (más simple)
		// return s.generateSyntacticFix("", errors.New(command.Errors[0]))
	}
	return "Comando sintácticamente incorrecto"
}

// ✅ CORREGIDO: Usar _ para parámetros intencionalmente no usados
func (s *MongoAnalyzerService) generateSemanticFix(_ *entities.MongoCommand, err error) string {
	errorMsg := err.Error()
	
	if strings.Contains(errorMsg, "nombre de la base de datos") {
		return "Usa un nombre válido para la base de datos (sin caracteres especiales)"
	}
	if strings.Contains(errorMsg, "nombre de la colección") {
		return "Usa un nombre válido para la colección (no puede empezar con '$')"
	}
	if strings.Contains(errorMsg, "documento a insertar") {
		return "El documento debe tener al menos un campo: { campo: valor }"
	}
	if strings.Contains(errorMsg, "filtro de actualización") {
		return "Especifica un filtro: { campo: valor }"
	}
	if strings.Contains(errorMsg, "operador válido") {
		return "Usa operadores como $set: { $set: { campo: nuevoValor } }"
	}
	
	return "Revisa la lógica del comando"
}