package memory

import (
	"fmt"
	"log"
)

// NewSubsystem generates a new subsystem graph
func NewSubsystem() *Subsystem {
	vertices := map[int]*Vertex{}
	lookup := map[string]int{}
	metrics := Metrics{ResourceSummary: map[string]Summary{}}

	// TODO need to add metadata onto vertices
	s := Subsystem{Vertices: vertices, lookup: lookup, Metrics: metrics}

	// Create a top level vertex for all clusters that will be added
	// Question: should the root be above the subsystems?
	s.AddNode("", "root", "root", 1, "")
	return &s
}

// AddNode (a physical node) as a vertex, return the vertex id
func (s *Subsystem) AddNode(
	clusterName, name, typ string,
	size int32,
	unit string,
) int {

	// Add resource to the metrics, indexed by the level (clusterName)
	// that we care about
	if clusterName != "" {
		s.Metrics.CountResource(clusterName, typ)
	}

	id := s.counter
	newEdges := map[int]*Edge{}
	s.Vertices[id] = &Vertex{
		Identifier: id,
		Edges:      newEdges,
		Size:       size,
		Type:       typ,
		Unit:       unit,
	}
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
func (s *Subsystem) AddEdge(src, dest int, weight int, relation string) error {

	// We shoudn't be added identifiers that don't exist...
	// TODO: do we want to count edges?
	_, ok := s.Vertices[src]
	if !ok {
		return fmt.Errorf("vertex with identifier %d does not exist", src)
	}
	_, ok = s.Vertices[dest]
	if !ok {
		return fmt.Errorf("vertex with identifier %d does not exist", dest)
	}

	// add edge src --> dest
	newEdge := Edge{Weight: weight, Vertex: s.Vertices[dest], Relation: relation}
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
