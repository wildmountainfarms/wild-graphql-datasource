package jsonnode

import "encoding/json"

type Node interface {
	sealed() // make this a sealed interface by having unexported methods

	DeepCopy() Node
	String() string
	Serialize() json.RawMessage
	Marshal() ([]byte, error) // extends Marshaler
}
