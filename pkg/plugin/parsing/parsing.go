package parsing

import (
	"errors"
	"fmt"
	"github.com/emirpasic/gods/v2/maps/linkedhashmap"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/wildmountainfarms/wild-graphql-datasource/pkg/plugin/parsing/framemap"
	"github.com/wildmountainfarms/wild-graphql-datasource/pkg/plugin/querymodel"
	"github.com/wildmountainfarms/wild-graphql-datasource/pkg/util/jsonnode"
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

func ParseData(graphQlResponseData *jsonnode.Object, parsingOption querymodel.ParsingOption) (data.Frames, error, ParseDataErrorType) {
	if len(parsingOption.DataPath) == 0 {
		return nil, errors.New("data path cannot be empty"), FRIENDLY_ERROR
	}
	split := strings.Split(parsingOption.DataPath, ".")

	var currentData *jsonnode.Object = graphQlResponseData
	for _, part := range split[:len(split)-1] {
		newData := currentData.Get(part)
		if newData == nil {
			return nil, errors.New(fmt.Sprintf("Part of data path: %s does not exist! dataPath: %s", part, parsingOption.DataPath)), FRIENDLY_ERROR
		}
		switch value := newData.(type) {
		case *jsonnode.Object:
			currentData = value
		default:
			return nil, errors.New(fmt.Sprintf("Part of data path: %s is not a nested object! dataPath: %s", part, parsingOption.DataPath)), FRIENDLY_ERROR
		}
	}
	// after this for loop, currentData should be an array if everything is going well
	finalData := currentData.Get(split[len(split)-1])
	if finalData == nil {
		return nil, errors.New(fmt.Sprintf("Final part of data path: %s does not exist! dataPath: %s", split[len(split)-1], parsingOption.DataPath)), FRIENDLY_ERROR
	}

	var dataArray []*jsonnode.Object
	switch value := finalData.(type) {
	case *jsonnode.Array:
		dataArray = make([]*jsonnode.Object, len(*value))
		for i, element := range *value {
			switch typedElement := element.(type) {
			case *jsonnode.Object:
				dataArray[i] = typedElement
			default:
				return nil, errors.New(fmt.Sprintf("One of the elements inside the data array is not an object! element: %d is of type: %v", i, reflect.TypeOf(element))), FRIENDLY_ERROR
			}
		}
	case *jsonnode.Object:
		// It's also valid if the final part of the data path refers to an object.
		//   The only downside of this is that it makes configuration errors harder to diagnose.
		dataArray = []*jsonnode.Object{
			value,
		}
	default:
		return nil, errors.New(fmt.Sprintf("Final part of data path: is not an array! dataPath: %s type of result: %v", parsingOption.DataPath, reflect.TypeOf(value))), FRIENDLY_ERROR
	}

	// We store a fieldMap inside of this frameMap.
	//   fieldMap is a map of keys to array of data points. Upon first initialization of a particular key's value,
	//   an array should be chosen corresponding to the first value of that key.
	//   Upon subsequent element insertion, if the type of the array does not match that elements type, an error is thrown.
	//   This error is never expected to occur because a correct GraphQL response should never have a particular field be of different types
	fm := framemap.New()

	//labelsToFieldMapMap := map[Labels]map[string]interface{}{}

	for _, dataElement := range dataArray {
		flatData := jsonnode.NewObject()
		flattenData(dataElement, "", flatData)
		labels, err := getLabelsFromFlatData(flatData, parsingOption)
		if err != nil {
			return nil, err, FRIENDLY_ERROR // getLabelsFromFlatData must always return a friendly error
		}
		var fieldMap, fieldMapExists = fm.Get(labels)
		if !fieldMapExists {
			fieldMap = linkedhashmap.New[string, any]()
			fm.Put(labels, fieldMap)
		}

		for _, key := range flatData.Keys() {
			value := flatData.Get(key)
			existingFieldValues, fieldValuesExist := fieldMap.Get(key)

			if key == parsingOption.TimePath {
				var timeValue time.Time
				switch typedValue := value.(type) {
				case *jsonnode.String:
					// TODO allow user to customize time format
					// Look at https://stackoverflow.com/questions/522251/whats-the-difference-between-iso-8601-and-rfc-3339-date-formats
					//   and also consider using time.RFC339Nano instead
					parsedTime, err := time.Parse(time.RFC3339, typedValue.String())
					if err != nil {
						return nil, errors.New(fmt.Sprintf("Time could not be parsed! Time: %s", typedValue)), FRIENDLY_ERROR
					}
					timeValue = parsedTime
				case *jsonnode.Number:
					epochMillis, err := typedValue.Int64()
					if err != nil {
						return nil, err, UNKNOWN_ERROR
					}
					timeValue = time.UnixMilli(epochMillis)
				default:
					// This case should never happen because we never expect other types to pop up here
					return nil, errors.New(fmt.Sprintf("Unsupported time type! Time: %s type: %v", typedValue, reflect.TypeOf(typedValue))), FRIENDLY_ERROR
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
				fieldMap.Put(key, fieldValues)
			} else {
				if fieldValuesExist {
					switch typedExistingFieldValues := existingFieldValues.(type) {
					case []float64:
						switch typedValue := value.(type) {
						case *jsonnode.Number:
							number, err := typedValue.Float64()
							if err != nil {
								return nil, err, UNKNOWN_ERROR
							}
							fieldMap.Put(key, append(typedExistingFieldValues, number))
						default:
							return nil, errors.New(fmt.Sprintf("Existing field values for key: %s is float64, but got value with type: %v", key, reflect.TypeOf(value))), FRIENDLY_ERROR
						}
					case []string:
						switch typedValue := value.(type) {
						case *jsonnode.String:
							fieldMap.Put(key, append(typedExistingFieldValues, typedValue.String()))
						default:
							return nil, errors.New(fmt.Sprintf("Existing field values for key: %s is string, but got value with type: %v", key, reflect.TypeOf(value))), FRIENDLY_ERROR
						}
					case []bool:
						switch typedValue := value.(type) {
						case *jsonnode.Boolean:
							fieldMap.Put(key, append(typedExistingFieldValues, typedValue.Bool()))
						default:
							return nil, errors.New(fmt.Sprintf("Existing field values for key: %s is bool, but got value with type: %v", key, reflect.TypeOf(value))), FRIENDLY_ERROR
						}
					default:
						return nil, errors.New(fmt.Sprintf("This error should never occur. The existing array for time field values is of the type: %v", reflect.TypeOf(existingFieldValues))), UNKNOWN_ERROR
					}
				} else {
					switch typedValue := value.(type) {
					case *jsonnode.Number:
						number, err := typedValue.Float64()
						if err != nil {
							return nil, err, UNKNOWN_ERROR
						}
						fieldMap.Put(key, []float64{number})
					case *jsonnode.String:
						fieldMap.Put(key, []string{typedValue.String()})
					case *jsonnode.Boolean:
						fieldMap.Put(key, []bool{typedValue.Bool()})
					case jsonnode.Null:
						// do nothing for null
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
func getLabelsFromFlatData(flatData *jsonnode.Object, parsingOption querymodel.ParsingOption) (data.Labels, error) {
	labels := map[string]string{}
	for _, labelOption := range parsingOption.LabelOptions {
		switch labelOption.Type {
		case querymodel.CONSTANT:
			labels[labelOption.Name] = labelOption.Value
		case querymodel.FIELD:
			fieldValue := flatData.Get(labelOption.Value)
			if fieldValue == nil {
				return nil, errors.New(fmt.Sprintf("Label option: %s could not be satisfied as key %s does not exist", labelOption.Name, labelOption.Value))
			}
			switch typedFieldValue := fieldValue.(type) {
			case *jsonnode.String:
				labels[labelOption.Name] = typedFieldValue.String()
			default:
				return nil, errors.New(fmt.Sprintf("Label option: %s could not be satisfied as key %s is not a string. It's type is %v", labelOption.Name, labelOption.Value, reflect.TypeOf(typedFieldValue)))
			}
		}
	}
	return labels, nil
}

func flattenArray(array *jsonnode.Array, prefix string, flattenedData *jsonnode.Object) {
	for key, value := range *array {
		flattenedData.Put(
			fmt.Sprintf("%s%d", prefix, key),
			value,
		)
	}
}

func flattenData(originalData *jsonnode.Object, prefix string, flattenedData *jsonnode.Object) {
	for _, key := range originalData.Keys() {
		value := originalData.Get(key)
		switch typedValue := value.(type) {
		case *jsonnode.Object:
			flattenData(typedValue, prefix+key+".", flattenedData)
		case *jsonnode.Array:
			flattenArray(typedValue, prefix+key+".", flattenedData)
		default:
			flattenedData.Put(prefix+key, typedValue)
		}
	}
}
