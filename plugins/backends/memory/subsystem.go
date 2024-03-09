package memory

import (
	"fmt"
	"log"

	"github.com/converged-computing/jsongraph-go/jsongraph/metadata"
	jgf "github.com/converged-computing/jsongraph-go/jsongraph/v2/graph"
)

// NewSubsystem generates a new subsystem graph
func NewSubsystem(name string) *Subsystem {
	vertices := map[int]*Vertex{}
	lookup := map[string]int{}
	metrics := Metrics{ResourceSummary: map[string]Summary{}}
	s := Subsystem{
		Vertices: vertices,
		Lookup:   lookup,
		Metrics:  metrics,
		Name:     name,
	}

	// Create a top level vertex for all clusters that will be added
	// Question: should the root be above the subsystems?
	s.AddNode("", name, name, 1, "", metadata.Metadata{}, false)
	return &s
}

// LoadSubsystemNodes into the graph
// For addition, we can have a two way pointer from the subsystem node TO
// the dominant node and then back:
// The pointer TO the dominant subsystem let's us find it to delete the opposing one
// The other one is used during the search to find the subsystem node
func (g *ClusterGraph) LoadSubsystemNodes(
	clusterName string,
	nodes *jgf.JsonGraph,
	subsystem string,
) error {

	// Get the dominant subsystem
	dom := g.DominantSubsystem()

	// Does the subsystem exist? Remember this is for across clusters
	_, ok := g.subsystem[subsystem]
	if !ok {
		ss := NewSubsystem(subsystem)
		g.subsystem[subsystem] = ss
	}

	ss, lookup, err := g.addNodes(clusterName, nodes, subsystem)
	if err != nil {
		return err
	}

	// Count dominant vertices references
	count := 0

	// Now add edges
	for _, edge := range nodes.Graph.Edges {

		// We are currently just saving one direction
		// Not the boy band.
		if edge.Relation != containsRelation {
			continue
		}

		// Two cases:
		// 1. the src is in the dominant subsystem
		// 2. The src is not, and both node are defined in the graph here
		subIdx1, ok1 := lookup[edge.Source]
		subIdx2, ok2 := lookup[edge.Target]

		// Case 1: both are in the subsystem graph
		if ok1 && ok2 {
			// This says "subsystem resource in node"
			ss.AddInternalEdge(subIdx1, subIdx2, 0, edge.Relation, subsystem)
		} else {

			// We need the namespaced name for the dom lookup
			lookupName := getNamespacedName(clusterName, edge.Source)

			// Case 2: the src is in the dominant subsystem
			domIdx, ok := dom.Lookup[lookupName]
			if !ok || !ok2 {
				return fmt.Errorf("edge %s->%s is not internal, and not connected to the dominant subsystem", edge.Source, edge.Target)
			}
			count += 1
			// Now add the link... the node exists in the subsystem but references a
			// different subsystem as the edge.
			// This says "dominant subsystem node conatains subsystem resource"
			err := dom.AddSubsystemEdge(domIdx, ss.Vertices[subIdx2], 0, edge.Relation, subsystem)
			if err != nil {
				return err
			}
		}
	}
	log.Printf("We have made an in memory graph (subsystem %s) with %d vertices, with %d connections to the dominant!", subsystem, ss.CountVertices(), count)
	g.subsystem[subsystem] = ss

	// Show metrics
	ss.Metrics.Show()
	return nil
}

// AddNode (a physical node) as a vertex, return the vertex id
func (s *Subsystem) AddNode(
	clusterName, lookupName, typ string,
	size int32,
	unit string,
	meta metadata.Metadata,
	countResource bool,
) int {

	// Add resource to the metrics, indexed by the level (clusterName)
	// We don't count the root "root" of a subsystem, typically
	if countResource {
		s.Metrics.CountResource(clusterName, typ)
	}

	id := s.counter
	newEdges := map[int]*Edge{}
	newSubsystems := map[string]map[int]*Edge{}
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
	relation, subsystem string,
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

	// Add the reference of the new edge here
	srcVertex.Edges[dest.Identifier] = &newEdge
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
