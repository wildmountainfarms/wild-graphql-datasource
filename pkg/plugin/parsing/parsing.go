package parsing

import (
	"errors"
	"fmt"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/wildmountainfarms/wild-graphql-datasource/pkg/plugin/querymodel"
	"reflect"
	"strings"
	"time"
)

// the purpose of this file is to parse JSON data with configuration from a ParsingOption

func ParseData(graphQlResponseData map[string]interface{}, parsingOption querymodel.ParsingOption) (*data.Frame, error) {
	if len(parsingOption.DataPath) == 0 {
		return nil, errors.New("data path cannot be empty")
	}
	split := strings.Split(parsingOption.DataPath, ".")

	var currentData map[string]interface{} = graphQlResponseData
	for _, part := range split[:len(split)-1] {
		newData, exists := currentData[part]
		if !exists {
			return nil, errors.New(fmt.Sprintf("Part of data path: %s does not exist! dataPath: %s", part, parsingOption.DataPath))
		}
		switch value := newData.(type) {
		case map[string]interface{}:
			currentData = value
		default:
			return nil, errors.New(fmt.Sprintf("Part of data path: %s is not a nested object! dataPath: %s", part, parsingOption.DataPath))
		}
	}
	// after this for loop, currentData should be an array if everything is going well
	finalData, finalDataExists := currentData[split[len(split)-1]]
	if !finalDataExists {
		return nil, errors.New(fmt.Sprintf("Final part of data path: %s does not exist! dataPath: %s", split[len(split)-1], parsingOption.DataPath))
	}

	var dataArray []map[string]interface{}
	switch value := finalData.(type) {
	case []interface{}:
		dataArray = make([]map[string]interface{}, len(value))
		for i, element := range value {
			switch typedElement := element.(type) {
			case map[string]interface{}:
				dataArray[i] = typedElement
			default:
				return nil, errors.New(fmt.Sprintf("One of the elements inside the data array is not an object! element: %d is of type: %v", i, reflect.TypeOf(element)))
			}
		}
	default:
		return nil, errors.New(fmt.Sprintf("Final part of data path: is not an array! dataPath: %s type of result: %v", parsingOption.DataPath, reflect.TypeOf(value)))
	}

	// fieldMap is a map of keys to array of data points. Upon first initialization of a particular key's value,
	//   an array should be chosen corresponding to the first value of that key.
	//   Upon subsequent element insertion, if the type of the array does not match that elements type, an error is thrown.
	//   This error is never expected to occur because a correct GraphQL response should never have a particular field be of different types
	fieldMap := map[string]interface{}{}

	for _, dataElement := range dataArray {
		flatData := map[string]interface{}{}
		flattenData(dataElement, "", flatData)
		for key, value := range flatData {
			existingFieldValues, fieldValuesExist := fieldMap[key]

			if key == parsingOption.TimePath {
				var timeValue time.Time
				switch valueValue := value.(type) {
				case string:
					// TODO allow user to customize time format
					// Look at https://stackoverflow.com/questions/522251/whats-the-difference-between-iso-8601-and-rfc-3339-date-formats
					//   and also consider using time.RFC339Nano instead
					parsedTime, err := time.Parse(time.RFC3339, valueValue)
					if err != nil {
						return nil, errors.New(fmt.Sprintf("Time could not be parsed! Time: %s", valueValue))
					}
					timeValue = parsedTime
				case float64:
					timeValue = time.UnixMilli(int64(valueValue))
				case bool:
					return nil, errors.New("time field is a bool")
				default:
					// This case should never happen because we never expect other types to pop up here
					return nil, errors.New(fmt.Sprintf("Unsupported time type! Time: %s type: %v", valueValue, reflect.TypeOf(valueValue)))
				}
				var fieldValues []time.Time
				if fieldValuesExist {
					switch typedExistingFieldValues := existingFieldValues.(type) {
					case []time.Time:
						fieldValues = typedExistingFieldValues
					default:
						return nil, errors.New(fmt.Sprintf("This error should never occur. The existing array for time field values is of the type: %v", reflect.TypeOf(existingFieldValues)))
					}
				} else {
					fieldValues = []time.Time{}
				}
				fieldValues = append(fieldValues, timeValue)
				fieldMap[key] = fieldValues
			} else {
				if fieldValuesExist {
					switch typedExistingFieldValues := existingFieldValues.(type) {
					case []float64:
						switch typedValue := value.(type) {
						case float64:
							fieldMap[key] = append(typedExistingFieldValues, typedValue)
						default:
							return nil, errors.New(fmt.Sprintf("Existing field values for key: %s is float64, but got value with type: %v", key, reflect.TypeOf(value)))
						}
					case []string:
						switch typedValue := value.(type) {
						case string:
							fieldMap[key] = append(typedExistingFieldValues, typedValue)
						default:
							return nil, errors.New(fmt.Sprintf("Existing field values for key: %s is string, but got value with type: %v", key, reflect.TypeOf(value)))
						}
					case []bool:
						switch typedValue := value.(type) {
						case bool:
							fieldMap[key] = append(typedExistingFieldValues, typedValue)
						default:
							return nil, errors.New(fmt.Sprintf("Existing field values for key: %s is bool, but got value with type: %v", key, reflect.TypeOf(value)))
						}
					default:
						return nil, errors.New(fmt.Sprintf("This error should never occur. The existing array for time field values is of the type: %v", reflect.TypeOf(existingFieldValues)))
					}
				} else {
					switch typedValue := value.(type) {
					case float64:
						fieldMap[key] = []float64{typedValue}
					case string:
						fieldMap[key] = []string{typedValue}
					case bool:
						fieldMap[key] = []bool{typedValue}
					default:
						return nil, errors.New(fmt.Sprintf("Unsupported and unexpected type for key: %s. Type is: %v", key, reflect.TypeOf(value)))
					}
				}
			}
		}
	}

	// create data frame response.
	// For an overview on data frames and how grafana handles them:
	//   https://grafana.com/developers/plugin-tools/introduction/data-frames
	// The goal here is to output a long format. If needed, prepare time series can transform it
	//   https://grafana.com/docs/grafana/latest/panels-visualizations/query-transform-data/transform-data/#prepare-time-series

	frame := data.NewFrame("response")

	for key, values := range fieldMap {
		frame.Fields = append(frame.Fields,
			data.NewField(key, nil, values),
		)
	}

	return frame, nil
}

func flattenArray[T interface{}](array []T, prefix string, flattenedData map[string]interface{}) {
	for key, value := range array {
		flattenedData[fmt.Sprintf("%s%d", prefix, key)] = value
	}
}

func flattenData(originalData map[string]interface{}, prefix string, flattenedData map[string]interface{}) {
	for key, value := range originalData {
		switch valueValue := value.(type) {
		case map[string]interface{}: // an object
			flattenData(valueValue, prefix+key+".", flattenedData)
		case []map[string]interface{}: // an array of objects
			for subKey, subValue := range valueValue {
				flattenData(subValue, fmt.Sprintf("%s%s.%d", prefix, key, subKey), flattenedData)
			}
		case []int:
			flattenArray(valueValue, prefix+key+".", flattenedData)
		case []int64:
			flattenArray(valueValue, prefix+key+".", flattenedData)
		case []float32:
			flattenArray(valueValue, prefix+key+".", flattenedData)
		case []float64:
			flattenArray(valueValue, prefix+key+".", flattenedData)
		case []bool:
			flattenArray(valueValue, prefix+key+".", flattenedData)
		case []uint:
			flattenArray(valueValue, prefix+key+".", flattenedData)
		case []uint64:
			flattenArray(valueValue, prefix+key+".", flattenedData)
		case []*int:
			flattenArray(valueValue, prefix+key+".", flattenedData)
		case []*int64:
			flattenArray(valueValue, prefix+key+".", flattenedData)
		case []*float32:
			flattenArray(valueValue, prefix+key+".", flattenedData)
		case []*float64:
			flattenArray(valueValue, prefix+key+".", flattenedData)
		case []*bool:
			flattenArray(valueValue, prefix+key+".", flattenedData)
		case []*uint:
			flattenArray(valueValue, prefix+key+".", flattenedData)
		case []*uint64:
			flattenArray(valueValue, prefix+key+".", flattenedData)
		default:
			flattenedData[prefix+key] = valueValue
		}
	}
}
