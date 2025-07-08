package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"mongo-analyzer/application/services"
	"mongo-analyzer/infrastructure/executor"
	"mongo-analyzer/infrastructure/lexer"
	"mongo-analyzer/infrastructure/parser"
	"mongo-analyzer/infrastructure/validator"
)

type AnalyzeRequest struct {
	Command string `json:"command"`
}

type AnalyzeResponse struct {
	IsValid         bool        `json:"is_valid"`
	Errors          []string    `json:"errors,omitempty"`
	TokenCount      int         `json:"token_count"`
	SuggestedFix    string      `json:"suggested_fix,omitempty"`
	ExecutionResult interface{} `json:"execution_result,omitempty"`
	ExecutionError  string      `json:"execution_error,omitempty"`
}

// ✅ CORS Middleware global
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Configurar headers CORS
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Length")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Si es un preflight request (OPTIONS), responder inmediatamente
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Continuar con el siguiente handler
		next.ServeHTTP(w, r)
	})
}

func main() {
	// Initialize dependencies
	lexer := lexer.NewMongoLexer()
	parser := parser.NewMongoParser()
	validator := validator.NewMongoValidator()
	executor := executor.NewMongoExecutor("mongodb://localhost:27017")
	
	analyzer := services.NewMongoAnalyzerService(lexer, parser, validator, executor)

	// Connect to MongoDB
	if err := executor.Connect(); err != nil {
		log.Printf("Warning: Could not connect to MongoDB: %v", err)
	}
	defer executor.Close()

	router := mux.NewRouter()
	
	// ✅ Aplicar middleware CORS a todas las rutas
	router.Use(corsMiddleware)
	
	// Rutas
	router.HandleFunc("/analyze", func(w http.ResponseWriter, r *http.Request) {
		handleAnalyze(w, r, analyzer)
	}).Methods("POST", "OPTIONS") // ✅ Agregar OPTIONS explícitamente

	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}).Methods("GET", "OPTIONS") // ✅ Agregar OPTIONS explícitamente

	fmt.Println("🚀 Servidor iniciado en puerto 8080")
	fmt.Println("🔍 Endpoint: POST /analyze")
	fmt.Println("💚 Health check: GET /health")
	fmt.Println("🌐 CORS habilitado para todos los orígenes")
	
	log.Fatal(http.ListenAndServe(":8080", router))
}

// ✅ Función simplificada (CORS ya manejado en middleware)
func handleAnalyze(w http.ResponseWriter, r *http.Request, analyzer *services.MongoAnalyzerService) {
	w.Header().Set("Content-Type", "application/json")

	var req AnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	result, err := analyzer.Analyze(req.Command)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := AnalyzeResponse{
		IsValid:         result.IsValid,
		Errors:          result.Errors,
		TokenCount:      result.TokenCount,
		SuggestedFix:    result.SuggestedFix,
		ExecutionResult: result.ExecutionResult,
	}

	if result.ExecutionError != nil {
		response.ExecutionError = result.ExecutionError.Error()
	}

	json.NewEncoder(w).Encode(response)
}