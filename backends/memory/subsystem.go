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

// DFSForMatch WILL is a depth first search for matches
// It starts by looking at total cluster resources on the top level,
// and then traverses into those that match the first check
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

	// Make a call on each of the top level resources
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
	if len(matches) != 0 {
		// Now that we got through the quicker check, do a deeper search
		return s.depthFirstSearch(matches, jobspec)
	}

	fmt.Println("  match: 😥️ no clusters could satisfy this request. We are sad")
	return matches, nil
}

// depthFirstSearch fully searches the graph finding a list of maches and a jobspec
func (s *Subsystem) depthFirstSearch(matches []string, jobspec *js.Jobspec) ([]string, error) {

	// Prepare a lookup of tasks for slots
	slots := map[string]*v1.Tasks{}
	for _, task := range jobspec.Tasks {
		slots[task.Slot] = &task
	}

	// Keep a list of final matches
	finalMatches := []string{}

	// Look through our potential matching clusters
	for _, cluster := range matches {
		fmt.Printf("\n  🔍️ Exploring cluster %s deeper with depth first search\n", cluster)

		// This is the root vertex of the cluster "cluster" we start with it
		root := s.lookup[cluster]
		vertex := s.Vertices[root]

		// Assume this is a match to start
		isMatch := true

		// Recursive function to recurse into slot resource and find count
		// of matches for the slot. This returns a count of the matching
		// slots under a parent level, recursing into child vertices until
		// we find the right type (and take a count) or keep exploring
		var findSlots func(vtx *Vertex, slot *v1.Resource) int32
		findSlots = func(vtx *Vertex, resource *v1.Resource) int32 {

			// This assumes the resource
			// Is the current vertex what we need? If yes, assess if it can satisfy
			slotsFound := int32(0)
			if vtx.Type == resource.Type {

				// I don't know if resource.Count can be zero, but be prepared...
				if resource.Count == 0 {
					return slotsFound
				}
				// How many full slots can we satisfy at this vertex?
				// TODO how to handle the slot per/total thing?
				return vtx.Size

			} else {

				// Otherwise, we haven't found the right level of the graph, keep going
				for _, child := range vtx.Edges {

					// Only interested in children. That sounds weird.
					if child.Relation == "contains" {
						slotsFound += findSlots(child.Vertex, resource)
					}
				}
			}
			return slotsFound
		}

		// Recursive function to Determine if a vertex satisfies a resource
		// Given a resource and a vertex root, it returns the count of vertices under
		// the root that satisfy the request.
		var satisfies func(vtx *Vertex, resource *v1.Resource, found int32) int32
		satisfies = func(vtx *Vertex, resource *v1.Resource, found int32) int32 {

			// All my life, searchin' for a vertex like youuu <3
			lookingAt := fmt.Sprintf("vertex '%s' (count=%d)", vtx.Type, vtx.Size)
			lookingFor := fmt.Sprintf("for '%s' (need=%d)", resource.Type, resource.Count)
			fmt.Printf("      => Checking %s %s\n", lookingAt, lookingFor)

			// A slot needs deeper exploration, and we need to add per_slot/total logic
			if resource.Type == "slot" {

				// Keep going until we have all the slots, or we run out of places to look
				return findSlots(vtx, resource)
			}

			// Wrong resource type, womp womp
			if vtx.Type != resource.Type {
				for _, child := range vtx.Edges {

					// Update our found count to include recursing all children
					if child.Relation == "contains" {
						found += satisfies(child.Vertex, resource, found)

						// Stop when we have enough
						if found >= resource.Count {
							return found
						}
					}
				}
			}

			// this resource type is satisfied, keep going and add to count
			if vtx.Type == resource.Type {
				return found + vtx.Size
			}

			// I'm not sure we'd ever get here, might want to check
			return found
		}

		// Traverse resource is the main function to handle traversing a cluster vertex
		var traverseResource = func(resource *v1.Resource) bool {
			fmt.Printf("\n    👀️ Looking for '%s' in cluster %s\n", resource.Type, cluster)

			// Case 1: A slot needs to be explored to determine if we can satsify
			// the count under it of some resource type
			if resource.Type == "slot" {

				// TODO: how does the slot Count (under tasks) fit in?
				// I don't understand what these counts are, because they seem like MPI tasks
				// but a slot can be defined at any level. So I'm going to ignore for now
				// Suggestion - this needs to be more clear in jobspec v2.
				// https://flux-framework.readthedocs.io/projects/flux-rfc/en/latest/spec_14.html

				// These are logical groups of "stuff" that need to be scheduled together
				slotsNeeded := resource.Count

				// Keep going until we have all the slots, or we run out of places to look
				slotsFound := int32(0)

				// This assumes that the slot value is defined in the next resource block
				if resource.With != nil {
					for _, subresource := range resource.With {
						slotsFound += findSlots(vertex, &subresource)

						// The slot is satisfied and we can continue searching resources
						if slotsFound >= slotsNeeded {
							return true
						}
					}
				}
				// The slot is satisfied and we can continue searching resources
				if slotsFound >= slotsNeeded {
					return true
				}
				return false

			} else {

				// Keep traversing vertices, start at the graph root
				foundMatches := satisfies(vertex, resource, int32(0))

				// We don't have a match, abort.
				if foundMatches < resource.Count {
					reason := fmt.Sprintf("%d of %s and %d needed\n", foundMatches, resource.Type, resource.Count)
					fmt.Printf("    ❌️ %s not a match, %s\n", cluster, reason)
					return false
				} else {
					reason := fmt.Sprintf("%d of needed %s satisfied", foundMatches, resource.Type)
					fmt.Printf("     ⏳️ %s still contender, %s\n", cluster, reason)
				}
			}
			// We get here if we assess a resource and vertex that isn't a slot, and foundMatches >= resource count
			return true
		}

		// Go through jobspec resources and determine satisfiability
		// This currently treats each item under resources separately
		// as opposed to one unit of work, and I'm not sure if that is
		// right. I haven't seen jobspecs in the wild with two entries
		// under resources.
		for _, resource := range jobspec.Resources {

			// Break out early if we can't sastify a resource group
			isMatch := traverseResource(&resource)
			if !isMatch {
				fmt.Printf("Resource %s is not a match for for cluster %s", resource.Label, cluster)
				break
			}

			// This is the recursive bit
			if resource.With != nil {
				for _, with := range resource.With {
					isMatch = traverseResource(&with)
					if !isMatch {
						break
					}
				}
			}
		}
		if isMatch {
			finalMatches = append(finalMatches, cluster)
		}

	}

	if len(finalMatches) == 0 {
		fmt.Println("    😥️ dfs: no clusters could satisfy this request. We are sad")
	} else {
		fmt.Printf("    🎯️ dfs: we found %d clusters to satisfy the request\n", len(finalMatches))
	}
	return finalMatches, nil
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
	srcVertex, ok := s.Vertices[src]
	if !ok {
		return fmt.Errorf("vertex with identifier %d does not exist", src)
	}
	destVertex, ok := s.Vertices[dest]
	if !ok {
		return fmt.Errorf("vertex with identifier %d does not exist", dest)
	}

	// add edge src --> dest
	newEdge := Edge{Weight: weight, Vertex: destVertex, Relation: relation}
	srcVertex.Edges[dest] = &newEdge
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
