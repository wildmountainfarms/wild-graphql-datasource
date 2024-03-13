package parsing

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"
)

func TestParsingJsonNumbersAsFloat64Bad(t *testing.T) {
	jsonString := `
{
  "a": 9223372036854775807,
  "b": 9223372036854775806.0
}
`
	jsonObject := map[string]interface{}{}

	err := json.Unmarshal([]byte(jsonString), &jsonObject)
	if err != nil {
		t.Error(err)
	}
	a, aExists := jsonObject["a"]
	b, bExists := jsonObject["b"]
	if !aExists || !bExists {
		t.Fatal("something does not exist")
	}
	//t.Log(fmt.Sprintf("Values of a: %d , b: %d", a, b))
	//t.Log(fmt.Sprintf("Types of a: %v , b: %v", reflect.TypeOf(a), reflect.TypeOf(b)))
	if reflect.TypeOf(a) != reflect.TypeOf(float64(0)) {
		t.Error("Type of a should be float64 (this is the default)")
	}
	if reflect.TypeOf(b) != reflect.TypeOf(float64(0)) {
		t.Error("Type of b should be float64 (this is the default)")
	}
	if a == 9223372036854775807 { // this is the maximum value of an int64, we show that it cannot be represented by a float
		t.Error("We don't expect a float64 to be able to represent the maximum int64 value")
	}
	if b == 9223372036854775806 {
		t.Error("We don't expect a float64 to be able to represent the maximum int64 value - 1")
	}
}

func TestParsingJsonNumbersAsNumbersGood(t *testing.T) {
	jsonString := `
{
  "a": 9223372036854775807,
  "b": 9223372036854775806.0
}
`
	jsonObject := map[string]interface{}{}

	d := json.NewDecoder(bytes.NewBufferString(jsonString))
	d.UseNumber()
	err := d.Decode(&jsonObject)
	if err != nil {
		t.Error(err)
	}
	a, aExists := jsonObject["a"]
	b, bExists := jsonObject["b"]
	if !aExists || !bExists {
		t.Fatal("something does not exist")
	}
	//t.Log(fmt.Sprintf("Values of a: %d , b: %d", a, b))
	//t.Log(fmt.Sprintf("Types of a: %v , b: %v", reflect.TypeOf(a), reflect.TypeOf(b)))

	if a != json.Number("9223372036854775807") {
		t.Error("We expect a to be represented precisely")
	}
	if b != json.Number("9223372036854775806.0") {
		t.Error("We expect b to be represented precisely")
	}
}
func TestParsingJsonArrayGivesArrayOfAny(t *testing.T) {
	jsonString := `
[
  51,
  32
]
`
	var jsonArray interface{}

	err := json.Unmarshal([]byte(jsonString), &jsonArray)
	if err != nil {
		t.Error(err)
	}
	if reflect.TypeOf(jsonArray) != reflect.TypeOf([]interface{}{}) {
		t.Error("We expect an array to be deserialized as []interface{}")
	}
}
