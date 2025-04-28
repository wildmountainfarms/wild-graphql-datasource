package querymodel

import (
	"github.com/emirpasic/gods/v2/sets"
	"github.com/emirpasic/gods/v2/sets/hashset"
)

// QueryModel represents data sent from the frontend to perform a query
type QueryModel struct {
	QueryText string `json:"queryText"`
	// The name of the operation, or a blank string to let the GraphQL server infer the operation name
	OperationName string `json:"operationName"`
	// The variables for the operation. May either be a string or a map[string]interface{} or nil
	Variables      interface{}     `json:"variables"`
	ParsingOptions []ParsingOption `json:"parsingOptions"`
}

type ParsingOption struct {
	// The path from the root to the array. This is dot-delimited
	DataPath          string   `json:"dataPath"`
	ExplodeArrayPaths []string `json:"explodeArrayPaths"`
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
	Name        string                  `json:"name"`
	Type        LabelOptionType         `json:"type"`
	Value       string                  `json:"value"`
	FieldConfig *LabelOptionFieldConfig `json:"fieldConfig"`
}
type LabelOptionFieldConfig struct {
	Required                  bool    `json:"required"`
	DefaultValue              *string `json:"defaultValue"`
	ExcludeFieldFromDataFrame *bool   `json:"excludeFieldFromDataFrame"`
}

func (parsingOption *ParsingOption) GetTimeField(key string) *TimeField {
	for _, timeField := range parsingOption.TimeFields {
		if timeField.TimePath == key {
			return &timeField
		}
	}
	return nil
}

// GetFieldsExcludedFromDataFrame returns all fields that should be excluded from the data frame.
// The current implementation determines this by only looking at the label options,
// but future changes to the behavior may change this.
func (parsingOption *ParsingOption) GetFieldsExcludedFromDataFrame() sets.Set[string] {
	// Note that we may consider returning a Set or map type in the future
	r := hashset.New[string]()
	for _, labelOption := range parsingOption.LabelOptions {
		if labelOption.Type == FIELD {
			fieldConfig := labelOption.FieldConfig
			if fieldConfig != nil {
				excludeFieldFromDataFrame := fieldConfig.ExcludeFieldFromDataFrame
				if excludeFieldFromDataFrame != nil && *excludeFieldFromDataFrame {
					r.Add(labelOption.Value)
				}
			}
		}
	}
	return r
}
