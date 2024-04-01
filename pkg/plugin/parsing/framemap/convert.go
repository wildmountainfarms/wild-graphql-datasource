package framemap

import (
	"encoding/json"
	"fmt"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/wildmountainfarms/wild-graphql-datasource/pkg/plugin/parsing/fieldsort"
	"time"
)

func (node *frameNode) getAllFields() []string {
	order := fieldsort.New()

	for _, row := range node.rows {
		order.State(row.FieldOrder)
	}

	return order.GetOrder()
}

func (node *frameNode) isTimeField(field string) bool {
	for _, row := range node.rows {
		if _, exists := row.TimeMap[field]; exists {
			return true
		}
		if _, exists := row.FieldMap[field]; exists {
			return false
		}
	}
	return false
}
func (node *frameNode) createField(field string) *data.Field {
	if node.isTimeField(field) {
		var values []*time.Time
		for _, row := range node.rows {
			value, exists := row.TimeMap[field]
			if exists {
				values = append(values, value)
			} else {
				values = append(values, nil)
			}
		}
		return data.NewField(field, node.labels, values)
	} else {
		var values []*json.RawMessage
		for _, row := range node.rows {
			value, exists := row.FieldMap[field]
			if exists {
				values = append(values, &value)
			} else {
				values = append(values, nil)
			}
		}
		return data.NewField(field, node.labels, values)
	}
}

func (f *FrameMap) ToFrames() []*data.Frame {
	// create data frame response.
	// For an overview on data frames and how grafana handles them:
	//   https://grafana.com/developers/plugin-tools/introduction/data-frames
	// The goal here is to output a long format. If needed, prepare time series can transform it
	//   https://grafana.com/docs/grafana/latest/panels-visualizations/query-transform-data/transform-data/#prepare-time-series

	// NOTE: The order of the frames here determines the order they appear in the legend in Grafana
	//   This is why we use a linkedhashmap.Map everywhere, as it maintains order.
	var r []*data.Frame
	frameMapIterator := f.data.Iterator()
	for frameMapIterator.Next() {
		node := frameMapIterator.Value()

		frameName := fmt.Sprintf("response %v", node.labels)
		frame := data.NewFrame(frameName)

		for _, field := range node.getAllFields() {
			frame.Fields = append(frame.Fields, node.createField(field))
		}
		r = append(r, frame)
	}
	return r
}
