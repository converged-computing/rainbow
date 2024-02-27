package memory

import (
	"fmt"
)

// NewSubsystem generates a new subsystem graph
func NewSubsystem() Subsystem {
	vertices := map[int]*Vertex{}
	return Subsystem{Vertices: vertices}
}

// AddNode (a physical node) as a vertex
func (s *Subsystem) AddNode(id int) {
	newEdges := map[int]*Edge{}
	s.Vertices[id] = &Vertex{Identifier: id, Edges: newEdges}
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
