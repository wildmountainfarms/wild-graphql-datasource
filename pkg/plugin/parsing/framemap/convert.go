package framemap

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/wildmountainfarms/wild-graphql-datasource/pkg/plugin/parsing/fieldsort"
	"github.com/wildmountainfarms/wild-graphql-datasource/pkg/util/jsonnode"
	"reflect"
	"time"
)

func (node *frameNode) getAllFields() []string {
	order := fieldsort.New()

	for _, row := range node.rows {
		order.State(row.FieldOrder)
	}

	return order.GetOrder()
}

func createFieldForNativeType[T comparable](node *frameNode, field string) *data.Field {
	var values []*T
	for _, row := range node.rows {
		rawValue, exists := row.FieldMap[field]
		if exists {
			switch rawValue.(type) {
			case jsonnode.Null:
				values = append(values, nil)
			default:
				value := rawValue.(T)
				values = append(values, &value)
			}
		} else {
			values = append(values, nil)
		}
	}
	return data.NewField(field, node.labels, values)
}
func createFieldForJsonNode(node *frameNode, field string) *data.Field {
	var values []*json.RawMessage
	for _, row := range node.rows {
		rawValue, exists := row.FieldMap[field]
		if exists {
			value := rawValue.(jsonnode.Node)
			serializedValue := value.Serialize()
			values = append(values, &serializedValue)
		} else {
			values = append(values, nil)
		}
	}
	return data.NewField(field, node.labels, values)
}

func (node *frameNode) createField(field string) *data.Field {
	var foundNull = false
	for _, row := range node.rows {
		if value, exists := row.FieldMap[field]; exists {
			switch value.(type) {
			case jsonnode.Null:
				foundNull = true
			case jsonnode.Number:
				return createFieldForJsonNode(node, field)
			case time.Time:
				return createFieldForNativeType[time.Time](node, field)
			case string:
				return createFieldForNativeType[string](node, field)
			case bool:
				return createFieldForNativeType[bool](node, field)
			case float64:
				return createFieldForNativeType[float64](node, field)
			default:
				message := fmt.Sprintf("unknown type: %v", reflect.TypeOf(value))
				log.DefaultLogger.Error(message)
				panic(errors.New(message))
			}
		}
	}
	if foundNull {
		return createFieldForJsonNode(node, field)
	}
	log.DefaultLogger.Info("Could not find field")
	panic(fmt.Errorf("could not find field: %s", field))
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
