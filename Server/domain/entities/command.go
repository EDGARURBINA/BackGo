package entities

type CommandType int

const (
	USE_DATABASE CommandType = iota
	CREATE_COLLECTION
	INSERT_ONE
	FIND
	UPDATE_ONE
	DELETE_ONE
	DROP_COLLECTION
	DROP_DATABASE
)

type MongoCommand struct {
	Type       CommandType
	Database   string
	Collection string
	Document   map[string]interface{}
	Filter     map[string]interface{}
	Update     map[string]interface{}
	IsValid    bool
	Errors     []string
	TokenCount int
}