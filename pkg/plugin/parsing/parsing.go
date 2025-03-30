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
	"strconv"
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

	var currentData jsonnode.Node = graphQlResponseData
	for i, part := range split {
		var newData jsonnode.Node
		switch typedCurrentData := currentData.(type) {
		case *jsonnode.Object:
			newData = typedCurrentData.Get(part)
			if newData == nil {
				return nil, errors.New(fmt.Sprintf("Part (index %d) of data path: %s does not exist! dataPath: %s", i, part, dataPath))
			}
		case *jsonnode.Array:
			partAsInteger, err := strconv.Atoi(part)
			if err != nil {
				return nil, errors.New(fmt.Sprintf("Part (index %d) of data path: %s should be an integer because the value before it was an array! dataPath: %s", i, part, dataPath))
			}
			if partAsInteger < 0 {
				return nil, errors.New(fmt.Sprintf("Part (index %d) of data path: %d should be a non-negative integer! dataPath: %s", i, partAsInteger, dataPath))
			}
			if partAsInteger >= len(*typedCurrentData) {
				return nil, errors.New(fmt.Sprintf("Part (index %d) of data path: %d must not fall out of bounds for array of length %d! dataPath: %s", i, partAsInteger, len(*typedCurrentData), dataPath))
			}
			newData = (*typedCurrentData)[partAsInteger]
			if newData == nil {
				return nil, errors.New(fmt.Sprintf("(Internal error) Part (index %d) of data path resulted in retrieving nil (this should never happen)! dataPath: %s", i, dataPath))
			}
		default:
			return nil, errors.New(fmt.Sprintf("(Internal error) Part (index %d) of data path resulted in unexpected currentData type! dataPath: %s, currentData type: %v", i, dataPath, reflect.TypeOf(currentData)))
		}
		if newData == nil {
			// We check whether newData is nil above, so this is just a sanity check and can never possibly occur
			panic(errors.New("newData is nil. This will never happen unless this function was incorrectly changed"))
		}
		switch value := newData.(type) {
		case *jsonnode.Object:
		case *jsonnode.Array:
		default:
			return nil, errors.New(fmt.Sprintf("Part (index %d) of data path: %s is not an object or array! Type is %v! dataPath: %s", i, part, reflect.TypeOf(value), dataPath))
		}
		currentData = newData
	}
	return currentData, nil
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

	expandedExplodeArrayPaths := expandPathsToSubPaths(parsingOption.ExplodeArrayPaths)

	for _, dataElement := range dataArray {
		flatDataExplodedArray := flattenAndExplode(dataElement, "", expandedExplodeArrayPaths)

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
				result := flattenAndExplode(typedValue, nestedArrayFullKey+".", explodeDataPaths)
				resultCrossed := crossObjects([]*jsonnode.Object{dataObject}, result)
				r = append(r, resultCrossed...)
			case *jsonnode.Array:
				innerArrayFullKey := nestedArrayFullKey + "._" // The "._" suffix is something we made up
				if slices.Contains(explodeDataPaths, innerArrayFullKey) {
					result := explodeArray([]*jsonnode.Object{jsonnode.NewObject()}, innerArrayFullKey, explodeDataPaths, typedValue)
					resultCrossed := crossObjects([]*jsonnode.Object{dataObject}, result)
					r = append(r, resultCrossed...)
				} else {
					flattenedData := jsonnode.NewObject()
					flattenArray(typedValue, innerArrayFullKey+".", flattenedData)
					newObject := dataObject.Clone()
					newObject.PutFrom(flattenedData)
					r = append(r, newObject)
				}
			default:
				newObject := dataObject.Clone()
				newObject.Put(nestedArrayFullKey, typedValue)
				r = append(r, newObject)
			}
		}
	}
	return r
}

func expandPathsToSubPaths(paths []string) []string {
	var r []string = nil
	for _, path := range paths {
		r = append(r, path)
		var subPath = path
		for {
			lastIndex := strings.LastIndex(subPath, ".")
			if lastIndex < 0 {
				break
			}
			subPath = subPath[:lastIndex]
			r = append(r, subPath)
		}
		strings.Split(path, ".")
	}
	return r
}

// flattenAndExplode will recursively flatten data and explode nested arrays if their path is contained within the explodeDataPaths argument.
//
// data is the source object to flatten.
// prefix is used to prefix the field names on each returned jsonnode.Object.
// explodeDataPaths is an array of data paths that point to a nested array.
// It is recommended to manually expand all paths into their subpaths as well, as this function does not check that super-paths are contained within explodeDataPaths.
func flattenAndExplode(data *jsonnode.Object, prefix string, explodeDataPaths []string) []*jsonnode.Object {
	var r = []*jsonnode.Object{
		jsonnode.NewObject(),
	}
	for _, key := range data.Keys() {
		value := data.Get(key)
		fullKey := prefix + key
		switch typedValue := value.(type) {
		case *jsonnode.Object:
			nestedDataArray := flattenAndExplode(typedValue, fullKey+".", explodeDataPaths)
			r = crossObjects(r, nestedDataArray)
		case *jsonnode.Array:
			if slices.Contains(explodeDataPaths, fullKey) {
				// Note that if value is an empty array, explodeArray() will return an empty array
				//   This could cause confusion when someone gets back an empty or mostly empty dataframe,
				//   but this is intended behavior.
				r = explodeArray(r, fullKey, explodeDataPaths, typedValue)
			} else {
				flattenedData := jsonnode.NewObject()
				flattenArray(typedValue, fullKey+".", flattenedData)
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
