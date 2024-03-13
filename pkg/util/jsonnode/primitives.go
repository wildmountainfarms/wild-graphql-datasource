package jsonnode

import (
	"encoding/json"
	"strconv"
)

type Number json.Number

func (_ Number) sealed() {}
func (n Number) String() string {
	return n.Number().String()
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

func (b Boolean) String() string {
	if b {
		return "true"
	}
	return "false"
}
func (b Boolean) Bool() bool {
	return bool(b)
}

type String string

func (_ String) sealed() {}

func (s String) String() string {
	return string(s)
}

type Null bool

const NULL Null = false

func (_ Null) sealed() {}

func (n Null) String() string {
	return "null"
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
