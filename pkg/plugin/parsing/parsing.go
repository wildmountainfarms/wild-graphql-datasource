package parsing

import (
	"errors"
	"fmt"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/wildmountainfarms/wild-graphql-datasource/pkg/plugin/parsing/framemap"
	"github.com/wildmountainfarms/wild-graphql-datasource/pkg/plugin/querymodel"
	"reflect"
	"strings"
	"time"
)

// the purpose of this file is to parse JSON data with configuration from a ParsingOption

type ParseDataErrorType int

const (
	NO_ERROR       ParseDataErrorType = 0
	FRIENDLY_ERROR ParseDataErrorType = 1
	UNKNOWN_ERROR  ParseDataErrorType = 2
)

func ParseData(graphQlResponseData map[string]interface{}, parsingOption querymodel.ParsingOption) (data.Frames, error, ParseDataErrorType) {
	if len(parsingOption.DataPath) == 0 {
		return nil, errors.New("data path cannot be empty"), FRIENDLY_ERROR
	}
	split := strings.Split(parsingOption.DataPath, ".")

	var currentData map[string]interface{} = graphQlResponseData
	for _, part := range split[:len(split)-1] {
		newData, exists := currentData[part]
		if !exists {
			return nil, errors.New(fmt.Sprintf("Part of data path: %s does not exist! dataPath: %s", part, parsingOption.DataPath)), FRIENDLY_ERROR
		}
		switch value := newData.(type) {
		case map[string]interface{}:
			currentData = value
		default:
			return nil, errors.New(fmt.Sprintf("Part of data path: %s is not a nested object! dataPath: %s", part, parsingOption.DataPath)), FRIENDLY_ERROR
		}
	}
	// after this for loop, currentData should be an array if everything is going well
	finalData, finalDataExists := currentData[split[len(split)-1]]
	if !finalDataExists {
		return nil, errors.New(fmt.Sprintf("Final part of data path: %s does not exist! dataPath: %s", split[len(split)-1], parsingOption.DataPath)), FRIENDLY_ERROR
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
				return nil, errors.New(fmt.Sprintf("One of the elements inside the data array is not an object! element: %d is of type: %v", i, reflect.TypeOf(element))), FRIENDLY_ERROR
			}
		}
	default:
		return nil, errors.New(fmt.Sprintf("Final part of data path: is not an array! dataPath: %s type of result: %v", parsingOption.DataPath, reflect.TypeOf(value))), FRIENDLY_ERROR
	}

	// We store a fieldMap inside of this frameMap.
	//   fieldMap is a map of keys to array of data points. Upon first initialization of a particular key's value,
	//   an array should be chosen corresponding to the first value of that key.
	//   Upon subsequent element insertion, if the type of the array does not match that elements type, an error is thrown.
	//   This error is never expected to occur because a correct GraphQL response should never have a particular field be of different types
	fm := framemap.CreateFrameMap()

	//labelsToFieldMapMap := map[Labels]map[string]interface{}{}

	for _, dataElement := range dataArray {
		flatData := map[string]interface{}{}
		flattenData(dataElement, "", flatData)
		labels, err := getLabelsFromFlatData(flatData, parsingOption)
		if err != nil {
			return nil, err, FRIENDLY_ERROR // getLabelsFromFlatData must always return a friendly error
		}
		var fieldMap, fieldMapExists = fm.Get(labels)
		if !fieldMapExists {
			fieldMap = map[string]interface{}{}
			fm.Put(labels, fieldMap)
		}

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
						return nil, errors.New(fmt.Sprintf("Time could not be parsed! Time: %s", valueValue)), FRIENDLY_ERROR
					}
					timeValue = parsedTime
				case float64:
					timeValue = time.UnixMilli(int64(valueValue))
				case bool:
					return nil, errors.New("time field is a bool"), FRIENDLY_ERROR
				default:
					// This case should never happen because we never expect other types to pop up here
					return nil, errors.New(fmt.Sprintf("Unsupported time type! Time: %s type: %v", valueValue, reflect.TypeOf(valueValue))), FRIENDLY_ERROR
				}
				var fieldValues []time.Time
				if fieldValuesExist {
					switch typedExistingFieldValues := existingFieldValues.(type) {
					case []time.Time:
						fieldValues = typedExistingFieldValues
					default:
						return nil, errors.New(fmt.Sprintf("This error should never occur. The existing array for time field values is of the type: %v", reflect.TypeOf(existingFieldValues))), UNKNOWN_ERROR
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
							return nil, errors.New(fmt.Sprintf("Existing field values for key: %s is float64, but got value with type: %v", key, reflect.TypeOf(value))), FRIENDLY_ERROR
						}
					case []string:
						switch typedValue := value.(type) {
						case string:
							fieldMap[key] = append(typedExistingFieldValues, typedValue)
						default:
							return nil, errors.New(fmt.Sprintf("Existing field values for key: %s is string, but got value with type: %v", key, reflect.TypeOf(value))), FRIENDLY_ERROR
						}
					case []bool:
						switch typedValue := value.(type) {
						case bool:
							fieldMap[key] = append(typedExistingFieldValues, typedValue)
						default:
							return nil, errors.New(fmt.Sprintf("Existing field values for key: %s is bool, but got value with type: %v", key, reflect.TypeOf(value))), FRIENDLY_ERROR
						}
					default:
						return nil, errors.New(fmt.Sprintf("This error should never occur. The existing array for time field values is of the type: %v", reflect.TypeOf(existingFieldValues))), UNKNOWN_ERROR
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
						return nil, errors.New(fmt.Sprintf("Unsupported and unexpected type for key: %s. Type is: %v", key, reflect.TypeOf(value))), UNKNOWN_ERROR
					}
				}
			}
		}
	}

	return fm.ToFrames(), nil, NO_ERROR
}

// Given flatData and label options, computes the labels or returns a friendly error
func getLabelsFromFlatData(flatData map[string]interface{}, parsingOption querymodel.ParsingOption) (data.Labels, error) {
	labels := map[string]string{}
	for _, labelOption := range parsingOption.LabelOptions {
		switch labelOption.Type {
		case querymodel.CONSTANT:
			labels[labelOption.Name] = labelOption.Value
		case querymodel.FIELD:
			fieldValue, fieldExists := flatData[labelOption.Value]
			if !fieldExists {
				return nil, errors.New(fmt.Sprintf("Label option: %s could not be satisfied as key %s does not exist", labelOption.Name, labelOption.Value))
			}
			switch typedFieldValue := fieldValue.(type) {
			case string:
				labels[labelOption.Name] = typedFieldValue
			default:
				return nil, errors.New(fmt.Sprintf("Label option: %s could not be satisfied as key %s is not a string. It's type is %v", labelOption.Name, labelOption.Value, reflect.TypeOf(typedFieldValue)))
			}
		}
	}
	return labels, nil
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
