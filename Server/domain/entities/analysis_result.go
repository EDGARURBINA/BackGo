package entities

type AnalysisResult struct {
	Command          *MongoCommand
	IsValid          bool
	Errors           []string
	TokenCount       int
	SuggestedFix     string
	ExecutionResult  interface{}
	ExecutionError   error
}