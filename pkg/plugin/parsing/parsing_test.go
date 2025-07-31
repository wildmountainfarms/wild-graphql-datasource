package parsing

import (
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"testing"
	"time"

	"github.com/wildmountainfarms/wild-graphql-datasource/pkg/util/jsonnode"
)

func TestExplodeArrayWhenGivenArrayOfNumbers(t *testing.T) {
	jsonString := `
[
  1,
  55,
  99
]`
	var array = jsonnode.NewArray()
	err := json.Unmarshal([]byte(jsonString), &array)
	if err != nil {
		t.Fatal("Could not unmarshal JSON", err)
	}
	results := explodeArray([]*jsonnode.Object{jsonnode.NewObject()}, "value", nil, array)
	if len(results) != 3 {
		t.Fatalf("len(results) is unexpected! its value is: %d", len(results))
	}
	for i, expectedResult := range []jsonnode.Number{
		jsonnode.Number("1"),
		jsonnode.Number("55"),
		jsonnode.Number("99"),
	} {
		if value, ok := results[i].Get("value").(jsonnode.Number); ok {
			if value != expectedResult {
				t.Errorf("Unexpected value on object %d! Got %s but expected %s", i, value, expectedResult)
			}
		} else {
			t.Errorf("object %d from results does not have a numeric value in its value field!", i)
		}
	}
}

func TestExplodeArrayWhenGivenArrayOfArrays(t *testing.T) {
	jsonString := `
[
  [1, 3],
  [55, 56],
  [99, 100]
]`
	var array = jsonnode.NewArray()
	err := json.Unmarshal([]byte(jsonString), &array)
	if err != nil {
		t.Fatal("Could not unmarshal JSON", err)
	}
	results := explodeArray([]*jsonnode.Object{jsonnode.NewObject()}, "value", nil, array)
	if len(results) != 3 {
		t.Fatalf("len(results) is unexpected! its value is: %d", len(results))
	}

	for i, expectedResultValues := range [][]jsonnode.Number{
		{jsonnode.Number("1"), jsonnode.Number("3")},
		{jsonnode.Number("55"), jsonnode.Number("56")},
		{jsonnode.Number("99"), jsonnode.Number("100")},
	} {
		for j, expectedResultValue := range expectedResultValues {
			if value, ok := results[i].Get(fmt.Sprintf("value._.%d", j)).(jsonnode.Number); ok {
				if value != expectedResultValue {
					t.Errorf("Unexpected value on %d . %d! Got %s but expected %s", i, j, value, expectedResultValue)
				}
			} else {
				t.Errorf("value %d . %d is not numeric!", i, j)
			}
		}
		//println(string(results[i].Serialize()))
	}
}

func TestExpandPathsToSubPaths(t *testing.T) {
	result1 := expandPathsToSubPaths([]string{
		"a.b.c",
		"c.b.a",
	})
	if !slices.Equal(result1, []string{
		"a.b.c",
		"a.b",
		"a",
		"c.b.a",
		"c.b",
		"c",
	}) {
		t.Errorf("Unexpected result from expandPathsToSubPaths. result1: %#v", result1)
	}
}

func TestFlattenAndExplode(t *testing.T) {
	jsonString := `
{
  "serverName": "awesome sauce",
  "data": [
    {
      "dateMillis": 1234,
      "node": "node 1",
      "data": [
        { "processor": 0, "temperature": 30.2 },
        { "processor": 1, "temperature": 31 }
      ]
    },
    {
      "dateMillis": 12345,
      "node": "node 2",
      "data": [
        { "processor": 0, "temperature": 35.2 },
        { "processor": 1, "temperature": 33.0 }
      ]
    }
  ]
}
`
	var object = jsonnode.NewObject()
	err := json.Unmarshal([]byte(jsonString), &object)
	if err != nil {
		t.Fatal("Could not unmarshal JSON", err)
	}
	resultsWithMisconfiguredPaths := flattenAndExplode(object, "", []string{"data.data"})
	if len(resultsWithMisconfiguredPaths) != 1 {
		t.Errorf("When we misconfigure the explodeDataPaths, we expect a length of 1, but got %d", len(resultsWithMisconfiguredPaths))
	}
	results := flattenAndExplode(object, "", []string{"data", "data.data"})
	if len(results) != 4 {
		t.Fatalf("Incorrect results size! size is %d", len(results))
	}
	for i, result := range results {
		serverNameNode := result.Get("serverName")
		if serverName, ok := serverNameNode.(jsonnode.String); ok {
			if serverName.String() != "awesome sauce" {
				t.Errorf("Element %d had unexpected serverName value: %s", i, serverName)
			}
		} else {
			t.Errorf("Unexpected type for element %d temperature. type: %v", i, reflect.TypeOf(serverNameNode))
		}
	}
	// Note that we are testing that the strings remain unchanged for the number type (some of these have a trailing ".0", others do not)
	for i, expectedTemperature := range []string{"30.2", "31", "35.2", "33.0"} {
		temperatureNode := results[i].Get("data.data.temperature")
		if temperature, ok := temperatureNode.(jsonnode.Number); ok {
			if temperature.String() != expectedTemperature {
				t.Errorf("Element %d had unexpected temperature value: %s", i, temperature)
			}
		} else {
			t.Errorf("Unexpected type for element %d temperature. type: %v", i, reflect.TypeOf(temperatureNode))
		}
	}
	//resultsArray := jsonnode.NewArray()
	//for _, result := range results {
	//	resultsArray.Add(result)
	//}
	//println(string(resultsArray.Serialize()))
}

func TestFlattenAndExplodeAnEmptyArray(t *testing.T) {
	jsonString := `
{
  "serverName": "awesome sauce",
  "data": [
  ]
}
`
	var object = jsonnode.NewObject()
	err := json.Unmarshal([]byte(jsonString), &object)
	if err != nil {
		t.Fatal("Could not unmarshal JSON", err)
	}
	results := flattenAndExplode(object, "", []string{"data", "data.data"})
	if len(results) != 0 {
		t.Fatalf("Incorrect results size! size is %d", len(results))
	}
}

func TestFlattenAndExplodeWithNestedArrayThatGetsFlattened(t *testing.T) {
	jsonString := `
{
  "serverName": "awesome sauce",
  "data": [
    [1, 2],
    [3, 4]
  ]
}
`
	var object = jsonnode.NewObject()
	err := json.Unmarshal([]byte(jsonString), &object)
	if err != nil {
		t.Fatal("Could not unmarshal JSON", err)
	}
	results := flattenAndExplode(object, "", []string{"data"})
	if len(results) != 2 {
		t.Fatalf("Incorrect results size! size is %d", len(results))
	}
	for i, expectedDataValues := range [][]jsonnode.Number{
		{jsonnode.Number("1"), jsonnode.Number("2")},
		{jsonnode.Number("3"), jsonnode.Number("4")},
	} {
		result := results[i]
		serverNameNode := result.Get("serverName")
		if serverName, ok := serverNameNode.(jsonnode.String); ok {
			if serverName.String() != "awesome sauce" {
				t.Errorf("Element %d had unexpected serverName value: %s", i, serverName)
			}
		} else {
			t.Errorf("Unexpected type for element %d temperature. type: %v", i, reflect.TypeOf(serverNameNode))
		}

		for j, expectedDataValue := range expectedDataValues {
			dataValue := result.Get(fmt.Sprintf("data._.%d", j))
			if dataValueNumber, ok := dataValue.(jsonnode.Number); ok {
				if dataValueNumber != expectedDataValue {
					t.Errorf("Result %d . nested data %d had unexpected value: %s", i, j, dataValue)
				}
			} else {
				t.Errorf("Unexpected type for result %d . nested data %d. type: %v", i, j, reflect.TypeOf(dataValue))
			}
		}
	}
}

func TestFlattenAndExplodeWithNestedArrayThatGetsExploded(t *testing.T) {
	jsonString := `
{
  "serverName": "awesome sauce",
  "data": [
    [1, 2],
    [3, 4]
  ]
}
`
	var object = jsonnode.NewObject()
	err := json.Unmarshal([]byte(jsonString), &object)
	if err != nil {
		t.Fatal("Could not unmarshal JSON", err)
	}
	results := flattenAndExplode(object, "", []string{"data", "data._"})
	if len(results) != 4 {
		t.Fatalf("Incorrect results size! size is %d", len(results))
	}
	for i, expectedDataValue := range []jsonnode.Number{
		jsonnode.Number("1"),
		jsonnode.Number("2"),
		jsonnode.Number("3"),
		jsonnode.Number("4"),
	} {
		result := results[i]
		serverNameNode := result.Get("serverName")
		if serverName, ok := serverNameNode.(jsonnode.String); ok {
			if serverName.String() != "awesome sauce" {
				t.Errorf("Element %d had unexpected serverName value: %s", i, serverName)
			}
		} else {
			t.Errorf("Unexpected type for element %d temperature. type: %v", i, reflect.TypeOf(serverNameNode))
		}

		dataValue := result.Get("data._")
		if dataValueNumber, ok := dataValue.(jsonnode.Number); ok {
			if dataValueNumber != expectedDataValue {
				t.Errorf("Result %d had unexpected value: %s", i, dataValue)
			}
		} else {
			t.Errorf("Unexpected type for result %d. type: %v", i, reflect.TypeOf(dataValue))
		}
	}
}

func TestParseTime(t *testing.T) {
	expectedTime := time.Date(2025, 05, 21, 0, 0, 0, 0, time.UTC)
	jsonInputs := []jsonnode.Node{
		jsonnode.NULL,
		jsonnode.Number("1747785600000"),
		jsonnode.String("1747785600000"),
		jsonnode.String("2025-05-21T00:00:00Z"),
		jsonnode.String("2025-05-21 00:00:00"),
	}

	for i, expectedValue := range []*time.Time{
		nil,
		&expectedTime,
		&expectedTime,
		&expectedTime,
	} {
		input := jsonInputs[i]

		dateValue, err, _ := parseTimeField(input)
		if err != nil {
			t.Errorf("Unexpected error at %d: %s", i, err)
		}

		if expectedValue == nil {
			if dateValue != nil {
				t.Errorf("Result %d had unexpected value: %s, expected: %s", i, dateValue, expectedValue)
			}
		} else if expectedValue == nil && dateValue != nil || *dateValue != *expectedValue {
			t.Errorf("Result %d had unexpected value: %s, expected: %s", i, dateValue, expectedValue)
		}
	}
}
