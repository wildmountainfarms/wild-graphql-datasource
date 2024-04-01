package fieldsort

type Order struct {
	data     *graph[string]
	vertices map[string]bool
}

func New() *Order {
	return &Order{
		data: &graph[string]{
			edges: make(map[string][]string),
		},
		vertices: make(map[string]bool),
	}
}

func (o *Order) State(ordering []string) {
	var last *string
	for _, element := range ordering {
		o.vertices[element] = true
		if last != nil {
			o.data.addEdge(*last, element)
		}

		elementCopy := element
		last = &elementCopy
	}
}
func (o *Order) GetOrder() []string {
	return o.data.topologicalSort(o.vertices)
}
