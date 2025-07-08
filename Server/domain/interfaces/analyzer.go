package interfaces

import "mongo-analyzer/domain/entities"

type CommandAnalyzer interface {
	Analyze(input string) (*entities.AnalysisResult, error)
}
