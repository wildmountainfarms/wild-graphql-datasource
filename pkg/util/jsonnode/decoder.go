package jsonnode

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

func createDecoder(r io.Reader) *json.Decoder {
	d := json.NewDecoder(r)
	d.UseNumber()
	return d
}

func decodeNode(token json.Token, decoder *json.Decoder) (Node, error) {
	{ // test if it's an array
		var nestedArray = Array{}
		err := nestedArray.decodeJSON(token, decoder)
		if err == nil {
			return &nestedArray, nil
		}
	}
	{ // test if it's an object
		var nestedObject = Object{}
		err := nestedObject.decodeJSON(token, decoder)
		if err == nil {
			return &nestedObject, nil
		}
	}
	primitiveNode := parsePrimitive(token)
	if primitiveNode != nil {
		return primitiveNode, nil
	}
	return nil, errors.New(fmt.Sprintf("Unknown token: %v", token))
}
