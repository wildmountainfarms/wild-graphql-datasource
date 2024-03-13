package jsonnode

import (
	"encoding/json"
	"testing"
)

func TestPrimitivesAndNestedArray(t *testing.T) {
	jsonString := `
[
  true,
  "asdf",
  null,
  1.1,
  [1, 2]
]
`
	var array = NewArray()
	//err := array.UnmarshalJSON([]byte(jsonString))
	err := json.Unmarshal([]byte(jsonString), &array)
	if err != nil {
		t.Fatal("Could not unmarshal JSON", err)
		return
	}
	if len(*array) == 0 {
		t.Error("Empty array")
		return
	}
	if value, ok := (*array)[0].(Boolean); !ok || value != true {
		if !ok {
			t.Error("First element is not a boolean")
			return
		}
		t.Error("First element's value is not true!")
		return
	}
	if value, ok := (*array)[1].(String); !ok || value != "asdf" {
		t.Error("Second element's value is not asdf!")
		return
	}
	if value, ok := (*array)[2].(Null); !ok || value != NULL {
		t.Error("Third element's value is not null!")
		return
	}
	if value, ok := (*array)[3].(Number); !ok || value.Number() != "1.1" {
		t.Error("Forth element's value is not 1.1!")
		return
	}
	{
		nestedArray, ok := (*array)[4].(*Array)
		if !ok {
			t.Error("Fifth element is not an array!")
			return
		}
		if len(*nestedArray) != 2 {
			t.Error("Incorrect data")
			return
		}
		if value, ok := (*nestedArray)[0].(Number); !ok || value.Number() != "1" {
			t.Error("Incorrect data")
			return
		}
		if value, ok := (*nestedArray)[1].(Number); !ok || value.Number() != "2" {
			t.Error("Incorrect data")
			return
		}
	}
}
func TestObject(t *testing.T) {
	jsonString := `
{
  "a": 1,
  "b": 2,
  "c": 3
}
`
	var object = NewObject()
	err := json.Unmarshal([]byte(jsonString), &object)
	if err != nil {
		t.Fatal("Could not unmarshal JSON", err)
		return
	}
	if len(object.Keys()) != 3 {
		t.Error("Incorrect keys length")
		return
	}
	{
		nodeValue := object.Get("a")
		if nodeValue == nil {
			t.Error("No value for key a")
			return
		}
		number, ok := nodeValue.(Number)
		if !ok {
			t.Error("Value is not a number")
			return
		}
		value, err := number.Int64()
		if err != nil {
			t.Fatal(err)
			return
		}
		if value != 1 {
			t.Error("Value is incorrect")
		}
	}
	{
		nodeValue := object.Get("b")
		if nodeValue == nil {
			t.Error("No value for key b")
			return
		}
		number, ok := nodeValue.(Number)
		if !ok {
			t.Error("Value is not a number")
			return
		}
		value, err := number.Int64()
		if err != nil {
			t.Fatal(err)
			return
		}
		if value != 2 {
			t.Error("Value is incorrect")
		}
	}
	{
		nodeValue := object.Get("c")
		if nodeValue == nil {
			t.Error("No value for key c")
			return
		}
		number, ok := nodeValue.(Number)
		if !ok {
			t.Error("Value is not a number")
			return
		}
		value, err := number.Int64()
		if err != nil {
			t.Fatal(err)
			return
		}
		if value != 3 {
			t.Error("Value is incorrect")
		}
	}

}

func TestVeryNestedData(t *testing.T) {
	jsonString := `

{
  "data": {
    "queryStatus": {
      "batteryVoltage": [
        {
          "dateMillis": 1704333209773,
          "packet": {
            "batteryVoltage": 27.8
          }
        }
      ]
    }
  }
}
`
	var object = NewObject()
	err := json.Unmarshal([]byte(jsonString), &object)
	if err != nil {
		t.Fatal("Could not unmarshal JSON", err)
		return
	}

}
