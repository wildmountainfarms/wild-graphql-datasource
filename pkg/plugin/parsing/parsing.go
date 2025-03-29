package parsing

import (
	"errors"
	"fmt"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/wildmountainfarms/wild-graphql-datasource/pkg/plugin/parsing/framemap"
	"github.com/wildmountainfarms/wild-graphql-datasource/pkg/plugin/querymodel"
	"github.com/wildmountainfarms/wild-graphql-datasource/pkg/util/jsonnode"
	"reflect"
	"slices"
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

func getNodeFromDataPath(graphQlResponseData *jsonnode.Object, dataPath string) (jsonnode.Node, error) {
	if len(dataPath) == 0 {
		return graphQlResponseData, nil
	}
	split := strings.Split(dataPath, ".")

	var currentData *jsonnode.Object = graphQlResponseData
	for _, part := range split[:len(split)-1] {
		newData := currentData.Get(part)
		if newData == nil {
			return nil, errors.New(fmt.Sprintf("Part of data path: %s does not exist! dataPath: %s", part, dataPath))
		}
		switch value := newData.(type) {
		case *jsonnode.Object:
			currentData = value
		default:
			// TODO if we come across an array, we should be able to index into it using the .<number> notation, like we do with fields
			return nil, errors.New(fmt.Sprintf("Part of data path: %s is not a nested object! Type is %v! dataPath: %s", part, reflect.TypeOf(value), dataPath))
		}
	}
	// after this for loop, currentData should be an array or an object if everything is going well
	finalData := currentData.Get(split[len(split)-1])
	if finalData == nil {
		return nil, errors.New(fmt.Sprintf("Final part of data path: %s does not exist! dataPath: %s", split[len(split)-1], dataPath))
	}
	return finalData, nil
}

func ParseData(graphQlResponseData *jsonnode.Object, parsingOption querymodel.ParsingOption) (data.Frames, error, ParseDataErrorType) {
	finalData, err := getNodeFromDataPath(graphQlResponseData, parsingOption.DataPath)
	if err != nil {
		return nil, err, FRIENDLY_ERROR
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
		return nil, errors.New(fmt.Sprintf("Final part of data path: is not an array or object! dataPath: %s type of result: %v", parsingOption.DataPath, reflect.TypeOf(value))), FRIENDLY_ERROR
	}

	// We store a fieldMap inside of this frameMap.
	//   fieldMap is a map of keys to array of data points. Upon first initialization of a particular key's value,
	//   an array should be chosen corresponding to the first value of that key.
	//   Upon subsequent element insertion, if the type of the array does not match that elements type, an error is thrown.
	//   This error is never expected to occur because a correct GraphQL response should never have a particular field be of different types
	fm := framemap.New()

	for _, dataElement := range dataArray {
		//theFlatData := jsonnode.NewObject()
		//flattenData(dataElement, "", theFlatData)
		flatDataExplodedArray := flattenOrExplode(dataElement, "", parsingOption.ExplodeArrayPaths)

		//for _, flatData := range []*jsonnode.Object{theFlatData} {
		for _, flatData := range flatDataExplodedArray {
			labels, err := getLabelsFromFlatData(flatData, parsingOption)
			if err != nil {
				return nil, err, FRIENDLY_ERROR // getLabelsFromFlatData must always return a friendly error
			}
			row := fm.NewRow(labels)
			row.FieldOrder = flatData.Keys()

			for _, key := range flatData.Keys() {
				value := flatData.Get(key)

				timeField := parsingOption.GetTimeField(key)
				if timeField != nil {
					var timePointer *time.Time
					switch typedValue := value.(type) {
					case jsonnode.String:
						// TODO allow user to customize time format
						// Look at https://stackoverflow.com/questions/522251/whats-the-difference-between-iso-8601-and-rfc-3339-date-formats
						//   and also consider using time.RFC339Nano instead
						parsedTime, err := time.Parse(time.RFC3339, typedValue.String())
						if err != nil {
							return nil, errors.New(fmt.Sprintf("Time could not be parsed! Time: %s", typedValue)), FRIENDLY_ERROR
						}
						timePointer = &parsedTime
					case jsonnode.Number:
						epochMillis, err := typedValue.Int64()
						if err != nil {
							return nil, err, UNKNOWN_ERROR
						}
						t := time.UnixMilli(epochMillis)
						timePointer = &t
					case jsonnode.Null:
						timePointer = nil
					default:
						// This case should never happen because we never expect other types to pop up here
						return nil, errors.New(fmt.Sprintf("Unsupported time type! Time: %s type: %v", typedValue, reflect.TypeOf(typedValue))), FRIENDLY_ERROR
					}
					if timePointer == nil {
						row.FieldMap[key] = jsonnode.NULL
					} else {
						row.FieldMap[key] = *timePointer
					}
				} else {
					switch typedValue := value.(type) {
					case jsonnode.String:
						row.FieldMap[key] = typedValue.String()
					case jsonnode.Boolean:
						row.FieldMap[key] = typedValue.Bool()
					case jsonnode.Number:
						parsedValue, err := typedValue.Float64()
						if err != nil {
							return nil, errors.New(fmt.Sprintf("Could not parse number: %s", typedValue.String())), UNKNOWN_ERROR
						}
						row.FieldMap[key] = parsedValue
						// NOTE: We are allowed to store a jsonnode.Number type directly into the FieldMap (it's part of the contract to support that),
						//   but we decide not to because alerting queries require float64s to be used
					case jsonnode.Null:
						row.FieldMap[key] = typedValue
					default:
						return nil, errors.New(fmt.Sprintf("Unsupported type! type: %v", reflect.TypeOf(typedValue))), UNKNOWN_ERROR
					}
				}
			}
		}
	}

	frames, err := fm.ToFrames()
	if err != nil {
		return nil, err, UNKNOWN_ERROR
	}
	return frames, nil, NO_ERROR
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
				fieldConfig := labelOption.FieldConfig
				if fieldConfig != nil && fieldConfig.Required {
					return nil, errors.New(fmt.Sprintf("Label option: %s could not be satisfied as key %s does not exist", labelOption.Name, labelOption.Value))
				} else if fieldConfig != nil && fieldConfig.DefaultValue != nil {
					labels[labelOption.Name] = *fieldConfig.DefaultValue
				}
				// else omit
			} else {
				switch typedFieldValue := fieldValue.(type) {
				case jsonnode.String:
					labels[labelOption.Name] = typedFieldValue.String()
				case jsonnode.Number:
					// TODO when we have a number that is a label, it will automatically be used by Grafana as a datapoint. Should we add logic to stop that from happening? (this todo comment isn't technically relevant to this specific area of the code)
					//   A potential solution is that maybe we should have an option to remove a field if it is being used as a label

					// TODO decide if we want to "normalize" when converting to string -- should 5.0 and 5 be the same string value?
					labels[labelOption.Name] = typedFieldValue.String()
				default:
					return nil, errors.New(fmt.Sprintf("Label option: %s could not be satisfied as key %s is not a string. It's type is %v", labelOption.Name, labelOption.Value, reflect.TypeOf(typedFieldValue)))
				}
			}
		}
	}
	return labels, nil
}

func flattenArray(array *jsonnode.Array, prefix string, flattenedData *jsonnode.Object) {
	for key, value := range *array {
		baseKey := fmt.Sprintf("%s%d", prefix, key)
		switch typedValue := value.(type) {
		case *jsonnode.Object:
			flattenData(typedValue, baseKey+".", flattenedData)
		case *jsonnode.Array:
			flattenArray(typedValue, baseKey+".", flattenedData)
		default:
			flattenedData.Put(
				baseKey,
				value,
			)
		}
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

func crossObjects(a []*jsonnode.Object, b []*jsonnode.Object) []*jsonnode.Object {
	slice := make([]*jsonnode.Object, len(a)*len(b))
	// TODO the ordering here determines the order of stuff in the dataframes. Consider making sure this is what we want
	for i, bObject := range b {
		for j, aObject := range a {
			newObject := jsonnode.NewObject()
			newObject.PutFrom(aObject)
			newObject.PutFrom(bObject)

			slice[j+i*len(a)] = newObject
		}
	}
	return slice
}
func explodeArray(data []*jsonnode.Object, nestedArrayFullKey string, explodeDataPaths []string, nestedArray *jsonnode.Array) []*jsonnode.Object {
	var r []*jsonnode.Object
	for _, nestedArrayElement := range *nestedArray {
		for _, dataObject := range data {
			switch typedValue := nestedArrayElement.(type) {
			case *jsonnode.Object:
				result := flattenOrExplode(typedValue, nestedArrayFullKey+".", explodeDataPaths)
				resultCrossed := crossObjects([]*jsonnode.Object{dataObject}, result)
				r = append(r, resultCrossed...)
			case *jsonnode.Array:
				// TODO an array nested within an array? Is that allowed?
				panic("TODO. This is not supported! an array within an array?? This probably isn't too bad to implement, but I don't feel like it rn")
			default:
				newObject := dataObject.Clone()
				newObject.Put(nestedArrayFullKey, typedValue)
				r = append(r, newObject)
			}
		}
	}
	return r
}

func flattenOrExplode(data *jsonnode.Object, prefix string, explodeDataPaths []string) []*jsonnode.Object {
	var r = []*jsonnode.Object{
		jsonnode.NewObject(),
	}
	for _, key := range data.Keys() {
		value := data.Get(key)
		fullKey := prefix + key
		switch typedValue := value.(type) {
		case *jsonnode.Object:
			nestedDataArray := flattenOrExplode(typedValue, fullKey+".", explodeDataPaths)
			r = crossObjects(r, nestedDataArray)
		case *jsonnode.Array:
			if slices.Contains(explodeDataPaths, fullKey) {
				r = explodeArray(r, fullKey, explodeDataPaths, typedValue)
			} else {
				flattenedData := jsonnode.NewObject()
				flattenArray(typedValue, prefix+key+".", flattenedData)
				for _, object := range r {
					object.PutFrom(flattenedData)
				}
			}
		default:
			for _, object := range r {
				object.Put(fullKey, typedValue)
			}
		}
	}
	return r
}
