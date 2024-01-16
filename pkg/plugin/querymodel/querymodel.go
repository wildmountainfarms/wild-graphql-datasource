package querymodel

// QueryModel represents data sent from the frontend to perform a query
type QueryModel struct {
	QueryText string `json:"queryText"`
	// The name of the operation, or a blank string to let the GraphQL server infer the operation name
	OperationName string `json:"operationName"`
	// The variables for the operation. May either be a string or a map[string]interface{}
	Variables      interface{}     `json:"variables"`
	ParsingOptions []ParsingOption `json:"parsingOptions"`
}

type ParsingOption struct {
	// The path from the root to the array. This is dot-delimited
	DataPath string `json:"dataPath"`
	// the time path relative to the data path.
	TimePath string `json:"timePath"`
}
