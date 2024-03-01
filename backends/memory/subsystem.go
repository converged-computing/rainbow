package memory

import (
	"fmt"
	"log"

	js "github.com/compspec/jobspec-go/pkg/jobspec/v1"
	v1 "github.com/compspec/jobspec-go/pkg/jobspec/v1"
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

// DFSForMatch WILL be a depth first search for matches
// Right now it's looking at total cluster resources on the top level,
// which is kind of terrible, but it's a start :)
func (s *Subsystem) DFSForMatch(jobspec *js.Jobspec) ([]string, error) {

	// Return a list of matching clusters
	matches := []string{}

	// Do a quick top level count for resource types
	totals := map[string]int32{}

	// Go sets loops to an initial value at start,
	// so we need a function to recurse into nested resources
	var checkResource func(resource *v1.Resource)
	checkResource = func(resource *v1.Resource) {
		count, ok := totals[resource.Type]
		if !ok {
			count = 0
		}
		count += resource.Count
		totals[resource.Type] = count

		// This is the recursive bit
		if resource.With != nil {
			for _, with := range resource.With {
				checkResource(&with)
			}
		}
	}

	for _, resource := range jobspec.Resources {
		checkResource(&resource)
	}

	// Compare against each cluster we know about
	for cluster, summary := range s.Metrics.ResourceSummary {
		fmt.Println(summary)

		isMatch := true
		for resourceType, needed := range totals {

			// TODO this should be part of a subsystem spec to ignore
			if resourceType == "slot" {
				continue
			}

			actual, ok := summary.Counts[resourceType]

			// We don't know. Assume we can't schedule
			if !ok {
				fmt.Printf("  ❌️ cluster %s is missing resource type %s, assuming cannot schedule\n", cluster, resourceType)
				isMatch = false
				break
			}
			// We don't have enough resources
			if int32(actual) < needed {
				fmt.Printf("  ❌️ cluster %s does not have sufficient resource type %s - actual %d vs needed %d\n", cluster, resourceType, actual, needed)
				isMatch = false
				break
			} else {
				fmt.Printf("  ✅️ cluster %s has sufficient resource type %s - actual %d vs needed %d\n", cluster, resourceType, actual, needed)
			}
		}
		// I don't think we need this, just be pedantic
		if isMatch {
			fmt.Printf("  match: 🎯️ cluster %s has enough resources and is a match\n", cluster)
			matches = append(matches, cluster)
		}
	}

	// No matches, womp womp.
	if len(matches) == 0 {
		fmt.Println("  match: 😥️ no clusters could satisfy this request. We are sad")
	}
	return matches, nil
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
