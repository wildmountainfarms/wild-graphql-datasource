package jsonnode

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type Array []Node

func (_ *Array) sealed() {}

func (a *Array) String() string {
	var stringArray []string
	for _, node := range *a {
		stringArray = append(stringArray, node.String())
	}
	return "Array [" + strings.Join(stringArray, ", ") + "]"
}

func (a *Array) decodeJSON(startToken json.Token, decoder *json.Decoder) error {
	if startToken != json.Delim('[') {
		return errors.New(fmt.Sprintf("Token is not the start of an array! Token: %v", startToken))
	}
	for { // inside the array
		token, err := decoder.Token()
		if err != nil {
			return err
		}
		if token == json.Delim(']') {
			return nil
		}
		node, err := decodeNode(token, decoder)
		if err != nil {
			return err
		}
		*a = append(*a, node)
	}
}

func (a *Array) UnmarshalJSON(data []byte) error {
	d := createDecoder(bytes.NewReader(data))

	var token, err = d.Token()
	if err != nil {
		return err
	}
	return a.decodeJSON(token, d)
}
