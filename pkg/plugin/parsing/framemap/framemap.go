package framemap

import (
	"github.com/emirpasic/gods/v2/maps/linkedhashmap"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

type frameNode struct {
	labels data.Labels
	rows   []*Row
}

func keyOfLabels(labels data.Labels) string {
	return labels.String()
}

type FrameMap struct {
	data *linkedhashmap.Map[string, *frameNode]
}

func New() *FrameMap {
	return &FrameMap{
		data: linkedhashmap.New[string, *frameNode](),
	}
}

func (f *FrameMap) getOrCreateFrameNode(labels data.Labels) *frameNode {
	mapKey := keyOfLabels(labels)
	values, exists := f.data.Get(mapKey)
	if exists {
		return values
	}
	node := &frameNode{
		labels: labels,
	}
	f.data.Put(
		mapKey,
		node,
	)
	return node
}
func (f *FrameMap) NewRow(labels data.Labels) *Row {
	node := f.getOrCreateFrameNode(labels)

	row := newRow()
	node.rows = append(node.rows, row)
	return row
}
