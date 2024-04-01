package framemap

type Row struct {
	FieldOrder []string
	// The field map. Supported types are as follows:
	//
	//  jsonnode.Number, jsonnode.Null
	//
	//  string, bool, float64, time.Time
	//
	// Please make sure you do not use pointer types.
	FieldMap map[string]any
}

func newRow() *Row {
	row := Row{
		FieldMap: map[string]any{},
	}
	return &row
}
