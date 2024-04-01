package fieldsort

// inspiration from https://reintech.io/blog/topological-sorting-in-go

type graph[Key comparable] struct {
	edges map[Key][]Key
}

func (g *graph[Key]) addEdge(u, v Key) {
	g.edges[u] = append(g.edges[u], v)
}

func (g *graph[Key]) topologicalSortUtil(v Key, visited map[Key]bool, stack *[]Key) {
	visited[v] = true

	for _, u := range g.edges[v] {
		if !visited[u] {
			g.topologicalSortUtil(u, visited, stack)
		}
	}

	*stack = append([]Key{v}, *stack...)
}

func (g *graph[Key]) topologicalSort(vertices map[Key]bool) []Key {
	var stack []Key
	visited := make(map[Key]bool)

	for vertex := range vertices {
		if !visited[vertex] {
			g.topologicalSortUtil(vertex, visited, &stack)
		}
	}

	return stack
}
