package jsonnode

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type Object struct {
	data  map[string]Node
	order []string
}

func NewObject() *Object {
	object := Object{
		data: map[string]Node{},
	}
	return &object
}

func (o *Object) decodeJSON(startToken json.Token, decoder *json.Decoder) error {
	if startToken != json.Delim('{') {
		return errors.New(fmt.Sprintf("Token is not the start of an object! Token: %v", startToken))
	}
	for {

		keyToken, err := decoder.Token()
		if err != nil {
			return err
		}
		if keyToken == json.Delim('}') {
			return nil
		}
		var key string
		switch typedToken := keyToken.(type) {
		case string:
			key = typedToken
		default:
			return errors.New(fmt.Sprintf("invalid token for key to object. token: %v", keyToken))
		}

		token, err := decoder.Token()
		if err != nil {
			return err
		}
		valueNode, err := decodeNode(token, decoder)
		if err != nil {
			return err
		}

		_, keyExists := o.data[key]
		if !keyExists {
			o.order = append(o.order, key)
		}
		o.data[key] = valueNode
	}
}

func (o *Object) UnmarshalJSON(data []byte) error {
	d := createDecoder(bytes.NewReader(data))

	var token, err = d.Token()
	if err != nil {
		return err
	}
	return o.decodeJSON(token, d)
}

func (_ *Object) sealed() {}

func (o *Object) String() string {
	var result []string
	for _, key := range o.Keys() {
		value := o.Get(key)
		result = append(result, fmt.Sprintf("%s : %s", key, value))
	}
	return strings.Join(result, ", ")
}

func (o *Object) Keys() []string {
	return o.order
}
func (o *Object) Get(key string) Node {
	node, exists := o.data[key]
	if !exists {
		return nil
	}
	return node
}
