package memory

import (
	"fmt"
	"log"
)

// NewSubsystem generates a new subsystem graph
func NewSubsystem() *Subsystem {
	vertices := map[int]*Vertex{}
	lookup := map[string]int{}
	metrics := Metrics{}

	// TODO need to add metadata onto vertices
	s := Subsystem{Vertices: vertices, lookup: lookup, Metrics: metrics}

	// Create a top level vertex for all clusters that will be added
	// Question: should the root be above the subsystems?
	s.AddNode("root")
	return &s
}

// AddNode (a physical node) as a vertex, return the vertex id
func (s *Subsystem) AddNode(name string) int {
	id := s.counter
	newEdges := map[int]*Edge{}
	s.Vertices[id] = &Vertex{Identifier: id, Edges: newEdges}
	s.counter += 1

	// If name is not null, we want to remember this node for later
	if name != "" {
		log.Printf("Adding special vertex %s at index %d\n", name, id)
		s.lookup[name] = id
	}
	return id
}

// GetNode returns the vertex if of a node, if it exists in the lookup
func (s *Subsystem) GetNode(name string) (int, bool) {
	id, ok := s.lookup[name]
	if ok {
		return id, true
	}
	return id, false
}

// Add an edge to the graph with a source and dest identifier
// Optionally add a weight. We aren't using this (but I think might)
func (s *Subsystem) AddEdge(src, dest int, weight int) error {

	// We shoudn't be added identifiers that don't exist...
	_, ok := s.Vertices[src]
	if !ok {
		return fmt.Errorf("vertex with identifier %d does not exist", src)
	}
	_, ok = s.Vertices[dest]
	if !ok {
		return fmt.Errorf("vertex with identifier %d does not exist", dest)
	}

	// add edge src --> dest
	newEdge := Edge{Weight: weight, Vertex: s.Vertices[dest]}
	s.Vertices[src].Edges[dest] = &newEdge
	return nil
}

// GetConnections to a vertex
func (s *Subsystem) GetConnections(src int) []int {

	conns := []int{}

	for _, edge := range s.Vertices[src].Edges {
		conns = append(conns, edge.Vertex.Identifier)
	}
	return conns
}

func (s *Subsystem) CountVertices() int {
	return len(s.Vertices)
}
