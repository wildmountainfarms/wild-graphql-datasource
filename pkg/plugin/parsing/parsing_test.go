package parsing

import (
	"encoding/json"
	"github.com/wildmountainfarms/wild-graphql-datasource/pkg/util/jsonnode"
	"reflect"
	"testing"
)

func TestFlattenOrExplode(t *testing.T) {
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
	resultsWithMisconfiguredPaths := flattenOrExplode(object, "", []string{"data.data"})
	if len(resultsWithMisconfiguredPaths) != 1 {
		t.Errorf("When we misconfigure the explodeDataPaths, we expect a length of 1, but got %d", len(resultsWithMisconfiguredPaths))
	}
	results := flattenOrExplode(object, "", []string{"data", "data.data"})
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
