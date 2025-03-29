package jsonnode

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type Number json.Number

func (_ Number) sealed() {}
func (n Number) DeepCopy() Node {
	return n
}
func (n Number) String() string {
	return n.Number().String()
}
func (n Number) Serialize() json.RawMessage {
	return json.RawMessage(n.Number())
}
func (n Number) Marshal() ([]byte, error) {
	return n.Serialize(), nil
}

func (n Number) Number() json.Number {
	return json.Number(n)
}

func (n Number) Float64() (float64, error) {
	return n.Number().Float64()
}

func (n Number) Int64() (int64, error) {
	return n.Number().Int64()
}
func (n Number) Uint64() (uint64, error) {
	return strconv.ParseUint(n.String(), 10, 64)
}

type Boolean bool

func (b Boolean) sealed() {}

func (b Boolean) DeepCopy() Node {
	return b
}
func (b Boolean) String() string {
	if b {
		return "true"
	}
	return "false"
}
func (b Boolean) Serialize() json.RawMessage {
	return json.RawMessage(b.String())
}
func (b Boolean) Marshal() ([]byte, error) {
	return b.Serialize(), nil
}
func (b Boolean) Bool() bool {
	return bool(b)
}

type String string

func (_ String) sealed() {}

func (s String) DeepCopy() Node {
	return s
}
func (s String) String() string {
	return string(s)
}
func (s String) Serialize() json.RawMessage {
	escaped, err := json.Marshal(s.String())
	if err != nil {
		panic(fmt.Errorf("error JSON-escaping string: %w", err))
	}
	return escaped
}
func (s String) Marshal() ([]byte, error) {
	return s.Serialize(), nil
}

type Null bool

const NULL Null = false

func (_ Null) sealed() {}

func (n Null) DeepCopy() Node {
	return n
}
func (n Null) String() string {
	return "null"
}
func (n Null) Serialize() json.RawMessage {
	return json.RawMessage("null")
}
func (n Null) Marshal() ([]byte, error) {
	return n.Serialize(), nil
}

func parsePrimitive(token json.Token) Node {
	switch typedToken := token.(type) {
	case json.Number:
		return Number(typedToken)
	case bool:
		return Boolean(typedToken)
	case string:
		return String(typedToken)
	case nil:
		return NULL
	}
	return nil
}
