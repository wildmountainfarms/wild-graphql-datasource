package jsonnode

type Node interface {
	sealed() // make this a sealed interface by having unexported methods

	String() string
	// json.Unmarshaler
	UnmarshalJSON(data []byte) error
}
