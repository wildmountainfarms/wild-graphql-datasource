package framemap

import (
	"fmt"
	"github.com/emirpasic/gods/v2/maps/linkedhashmap"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

type FrameAndLabels struct {
	labels data.Labels
	// A map of field names to an array of the values of that given column
	fieldMap *linkedhashmap.Map[string, any]
}

func keyOfLabels(labels data.Labels) string {
	return labels.String()
}

type FrameMap struct {
	data *linkedhashmap.Map[string, FrameAndLabels]
}

func New() *FrameMap {
	return &FrameMap{
		data: linkedhashmap.New[string, FrameAndLabels](),
	}
}

func (f *FrameMap) Get(labels data.Labels) (*linkedhashmap.Map[string, any], bool) {
	mapKey := keyOfLabels(labels)
	values, exists := f.data.Get(mapKey)
	if !exists {
		return nil, false
	}
	return values.fieldMap, true
}
func (f *FrameMap) Put(labels data.Labels, fieldMap *linkedhashmap.Map[string, any]) {
	mapKey := keyOfLabels(labels)
	f.data.Put(
		mapKey,
		FrameAndLabels{
			labels:   labels,
			fieldMap: fieldMap,
		},
	)
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
		frameAndLabels := frameMapIterator.Value()

		frameName := fmt.Sprintf("response %v", frameAndLabels.labels)
		frame := data.NewFrame(frameName)
		fieldMapIterator := frameAndLabels.fieldMap.Iterator()
		for fieldMapIterator.Next() {
			key := fieldMapIterator.Key()
			values := fieldMapIterator.Value()
			frame.Fields = append(frame.Fields,
				data.NewField(key, frameAndLabels.labels, values),
			)
		}
		r = append(r, frame)
	}
	return r
}
