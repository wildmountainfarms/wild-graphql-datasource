package plugin

import (
	"context"
	"testing"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

func TestQueryData(t *testing.T) {
	ds := Datasource{}

	resp, err := ds.QueryData(
		context.Background(),
		&backend.QueryDataRequest{
			Queries: []backend.DataQuery{
				{RefID: "A"},
			},
		},
	)
	if err != nil {
		t.Error(err)
	}

	if len(resp.Responses) != 1 {
		t.Fatal("QueryData must return a response")
	}
}

func TestParsingJsonNestedArray(t *testing.T) {
	const jsonString = `
{
    "data": {
        "query1": {
            "value1": 1,
            "value2": 0,
            "sub_value": [
                {
                    "value1": "string",
                    "value2": 1,
                    "sub_value": [
                        {
                            "value1": "string",
                            "value2": 1
                        }
                    ]
                }
            ]
        }
    }
}
`
	const inputJson = `
{
	"queryText": "query($from: String!, $to: String!) {\n query1(startTime: $from, endTime: $to, msecFormat: true) {\n value1\n\t value2\n sub_value {\n value1\n\t value2\n sub_value {\n        value1\n        value2\n      }\n    }\n  }\n}\n",
	"operationName": "",
	"variables": {"from": "1729241695217", "to": "1737017695217" },
	"parsingOptions": [
        {
          "dataPath": "query1",
          "timeFields": []
        }
      ]
}
	`

	const expectedNestedArrayValue = `{"schema":{"name":"response ","fields":[{"name":"value1","type":"number","typeInfo":{"frame":"float64","nullable":true},"labels":{}},{"name":"value2","type":"number","typeInfo":{"frame":"float64","nullable":true},"labels":{}},{"name":"sub_value","type":"other","typeInfo":{"frame":"json.RawMessage","nullable":true},"labels":{}}]},"data":{"values":[[1],[0],[[{"value1":"string","value2":1,"sub_value":[{"value1":"string","value2":1}]}]]]}}`

	ds, err := NewMockDatasource([]byte(jsonString))
	if err != nil {
		t.Fatal(err)
	}

	resp, err := ds.QueryData(
		context.Background(),
		&backend.QueryDataRequest{
			Queries: []backend.DataQuery{
				{
					RefID:         "A",
					MaxDataPoints: 970,
					Interval:      time.Millisecond * 7200000,
					JSON:          []byte(inputJson),
				},
			},
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	if len(resp.Responses) != 1 {
		t.Fatal("QueryData must return a response")
	}

	for _, r := range resp.Responses {
		if r.Error != nil {
			t.Fatal(r.Error)
		}

		for _, frame := range r.Frames {
			data, err := frame.MarshalJSON()
			if err != nil {
				t.Fatal(err)
			}

			if string(data) != expectedNestedArrayValue {
				t.Errorf("Unexpected JSON frame data: %s", string(data))
			}
		}
	}
}
