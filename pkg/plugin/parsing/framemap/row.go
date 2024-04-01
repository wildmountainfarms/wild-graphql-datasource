package framemap

import (
	"encoding/json"
	"time"
)

type Row struct {
	FieldOrder []string
	FieldMap   map[string]json.RawMessage
	TimeMap    map[string]*time.Time
}

func newRow() *Row {
	row := Row{
		FieldMap: map[string]json.RawMessage{},
		TimeMap:  map[string]*time.Time{},
	}
	return &row
}
