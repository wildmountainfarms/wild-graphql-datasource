package jsonnode

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)

type Number struct {
	number json.Number
}

func (_ *Number) sealed() {}
func (n *Number) String() string {
	return n.number.String()
}

func (n *Number) UnmarshalJSON(data []byte) error {
	d := json.NewDecoder(bytes.NewReader(data))
	d.UseNumber()
	var number json.Number
	err := d.Decode(&number)
	if err != nil {
		return err
	}
	n.number = number
	return nil
}

func (n *Number) Number() json.Number {
	return n.number
}

func (n *Number) Float64() (float64, error) {
	return n.number.Float64()
}

func (n *Number) Int64() (int64, error) {
	return n.number.Int64()
}
func (n *Number) Int() (int, error) {
	return strconv.Atoi(string(n.number))
}

type Boolean bool

func (b *Boolean) sealed() {}

func (b *Boolean) String() string {
	if *b {
		return "true"
	}
	return "false"
}
func (b *Boolean) UnmarshalJSON(data []byte) error {
	var value bool
	err := json.Unmarshal(data, &value)
	if err != nil {
		return err
	}
	*b = Boolean(value)
	return nil
}
func (b *Boolean) Bool() bool {
	if *b {
		return true
	}
	return false
}

type String string

func (_ *String) sealed() {}

func (s *String) String() string {
	return string(*s)
}
func (s *String) UnmarshalJSON(data []byte) error {
	var value string
	err := json.Unmarshal(data, &value)
	if err != nil {
		return err
	}
	*s = String(value)
	return nil
}

type Null bool

const NULL Null = false

func (_ Null) sealed() {}

func (n Null) String() string {
	return "null"
}
func (n Null) UnmarshalJSON(data []byte) error {
	d := createDecoder(bytes.NewReader(data))
	token, err := d.Token()
	if err != nil {
		return err
	}
	if token != nil {
		return errors.New(fmt.Sprintf("token is not a null token. token: %v", token))
	}
	return nil
}

func parsePrimitive(token json.Token) Node {
	switch typedToken := token.(type) {
	case json.Number:
		number := Number{typedToken}
		var node Node = &number
		return node
	case bool:
		value := Boolean(typedToken)
		var node Node = &value
		return node
	case string:
		value := String(typedToken)
		var node Node = &value
		return node
	case nil:
		return NULL
	}
	return nil
}
