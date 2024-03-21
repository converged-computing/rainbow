package memory

import (
	"fmt"

	"github.com/converged-computing/jsongraph-go/jsongraph/metadata"
)

// NewSubsystem generates a new subsystem graph
func NewSubsystem(name string) *Subsystem {
	vertices := map[int]*Vertex{}
	lookup := map[string]int{}
	metrics := Metrics{ResourceCounts: map[string]int64{}, Name: name}
	s := Subsystem{
		Vertices: vertices,
		Lookup:   lookup,
		Metrics:  metrics,
		Name:     name,
	}

	// Create a top level vertex for all clusters that will be added
	// Question: should the root be above the subsystems?
	s.AddNode("", name, 1, "", metadata.Metadata{}, false)
	return &s
}

// AddNode (a physical node) as a vertex, return the vertex id
func (s *Subsystem) AddNode(
	lookupName, typ string,
	size int32,
	unit string,
	meta metadata.Metadata,
	countResource bool,
) int {

	// Add resource to the metrics, indexed by the level (clusterName)
	// We don't count the root "root" of a subsystem, typically
	if countResource {
		s.Metrics.CountResource(typ)
	}

	id := s.counter
	newEdges := map[int]*Edge{}
	newSubsystems := map[string]map[int]*Edge{}

	// Add the subsystem node
	s.Vertices[id] = &Vertex{
		Identifier: id,
		Edges:      newEdges,
		Size:       size,
		Type:       typ,
		Unit:       unit,
		Metadata:   meta,
		Subsystems: newSubsystems,
	}
	s.counter += 1

	// If name is not null, we want to remember this node for later
	if lookupName != "" {
		s.Lookup[lookupName] = id
	}
	return id
}

// GetNode returns the vertex if of a node, if it exists in the lookup
func (s *Subsystem) GetNode(name string) (int, bool) {
	id, ok := s.Lookup[name]
	if ok {
		return id, true
	}
	return id, false
}

// Add an edge to the graph with a source and dest identifier
// This assumes they belong in the same subsystem (src subsystem == dest subsystem)
// Optionally add a weight. We aren't using this (but I think might)
func (s *Subsystem) AddInternalEdge(
	src, dest, weight int,
	relation,
	subsystem string,
) error {

	// We shoudn't be added identifiers that don't exist...
	srcVertex, ok := s.Vertices[src]
	if !ok {
		return fmt.Errorf("vertex with identifier %d does not exist", src)
	}
	destVertex, ok := s.Vertices[dest]
	if !ok {
		return fmt.Errorf("vertex with identifier %d does not exist", dest)
	}

	// add edge src --> dest
	// Right now subsystem references the source
	newEdge := Edge{
		Weight:    weight,
		Vertex:    destVertex,
		Relation:  relation,
		Subsystem: subsystem,
	}
	srcVertex.Edges[dest] = &newEdge
	s.Vertices[src] = srcVertex
	return nil
}

// Add an subsystem edge, meaning adding the edge AND a link to the dominant subsystem
// This would be called by the dominant to add an edge to itself
func (s *Subsystem) AddSubsystemEdge(
	src int,
	dest *Vertex,
	weight int,
	relation string,
	subsystem string,
) error {

	// The source vertex is owned by this subsystem
	// But we don't check dest, it's part of another subsystem
	srcVertex, ok := s.Vertices[src]
	if !ok {
		return fmt.Errorf("vertex with identifier %d does not exist", src)
	}

	// add edge src --> dest
	// Right now subsystem references the source
	newEdge := Edge{
		Weight:    weight,
		Vertex:    dest,
		Relation:  relation,
		Subsystem: subsystem,
	}

	// Add the reference of the new edge here, note to
	// subsystem edges. This makes the search easier so we don't
	// iterate thrugh subsystem AND dominant subsystem nodes.
	subsysEdges, ok := srcVertex.Subsystems[subsystem]
	if !ok {
		subsysEdges = map[int]*Edge{}
		srcVertex.Subsystems[subsystem] = subsysEdges
	}
	srcVertex.Subsystems[subsystem][dest.Identifier] = &newEdge
	s.Vertices[src] = srcVertex
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
