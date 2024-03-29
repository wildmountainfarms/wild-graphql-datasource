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
	TimeFields   []TimeField   `json:"timeFields"`
	LabelOptions []LabelOption `json:"labelOptions"`
}

type TimeField struct {
	TimePath string `json:"timePath"`

	// We will put time format options here as the frontend implements it
}

type LabelOptionType string

const (
	CONSTANT LabelOptionType = "constant"
	FIELD    LabelOptionType = "field"
)

type LabelOption struct {
	Name  string          `json:"name"`
	Type  LabelOptionType `json:"type"`
	Value string          `json:"value"`
}

func (parsingOption *ParsingOption) GetTimeField(key string) *TimeField {
	for _, timeField := range parsingOption.TimeFields {
		if timeField.TimePath == key {
			return &timeField
		}
	}
	return nil
}
