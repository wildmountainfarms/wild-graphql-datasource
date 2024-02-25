package framemap

import (
	"fmt"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"hash/fnv"
	"slices"
	"sort"
)

type FrameAndLabels struct {
	labels   data.Labels
	fieldMap map[string]interface{}
}

func labelsEqual(labelsA data.Labels, labelsB data.Labels) bool {
	if len(labelsA) != len(labelsB) {
		return false
	}
	for key, value := range labelsA {
		otherValue, exists := labelsB[key]
		if !exists || value != otherValue {
			return false
		}
	}
	return true
}
func labelsHash(labels data.Labels) uint64 {
	h := fnv.New64a()
	// remember, iteration over entries in a map in go is not defined! (That's dumb! Why are you like this Go!)
	keys := make([]string, 0, len(labels))
	for key := range labels {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	for _, key := range keys {
		_, err := h.Write([]byte(key))
		if err != nil {
			panic(fmt.Sprintf("Error writing to hash: %v", err))
		}
		_, err = h.Write([]byte(labels[key]))
		if err != nil {
			panic(fmt.Sprintf("Error writing to hash: %v", err))
		}
	}
	return h.Sum64()
}

type FrameMap struct {
	data map[uint64][]FrameAndLabels
}

func CreateFrameMap() FrameMap {
	return FrameMap{
		data: map[uint64][]FrameAndLabels{},
	}
}

func (f *FrameMap) Get(labels data.Labels) (map[string]interface{}, bool) {
	hash := labelsHash(labels)
	values, exists := f.data[hash]
	if !exists {
		return nil, false
	}

	for _, value := range values {
		if labelsEqual(value.labels, labels) {
			return value.fieldMap, true
		}
	}
	return nil, false
}
func (f *FrameMap) Put(labels data.Labels, fieldMap map[string]interface{}) {
	hash := labelsHash(labels)
	values, exists := f.data[hash]
	if !exists {
		f.data[hash] = []FrameAndLabels{{labels: labels, fieldMap: fieldMap}}
		return
	}
	for index, value := range values {
		if labelsEqual(value.labels, labels) {
			values[index] = FrameAndLabels{labels: labels, fieldMap: fieldMap}
			return
		}
	}

	newValues := append(values, FrameAndLabels{labels: labels, fieldMap: fieldMap})
	f.data[hash] = newValues
}

func (f *FrameMap) ToFrames() []*data.Frame {
	// create data frame response.
	// For an overview on data frames and how grafana handles them:
	//   https://grafana.com/developers/plugin-tools/introduction/data-frames
	// The goal here is to output a long format. If needed, prepare time series can transform it
	//   https://grafana.com/docs/grafana/latest/panels-visualizations/query-transform-data/transform-data/#prepare-time-series

	// NOTE: The order of the frames here determines the order they appear in the legend in Grafana
	// 	 A workaround on the frontend is to make the legend in "Table" mode and then sort the "Name" column: https://github.com/grafana/grafana/pull/69490
	var r []*data.Frame
	for _, values := range f.data {
		for _, value := range values {
			frameName := fmt.Sprintf("response %v", value.labels)
			frame := data.NewFrame(frameName)
			for key, values := range value.fieldMap {
				frame.Fields = append(frame.Fields,
					data.NewField(key, value.labels, values),
				)
			}
			r = append(r, frame)
		}
	}
	sort.Slice(r, func(i, j int) bool {
		return r[i].Name < r[j].Name // sort alphabetically by name
	})
	return r
}
