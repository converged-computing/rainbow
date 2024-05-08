package memory

import (
	"fmt"

	v1 "github.com/compspec/jobspec-go/pkg/nextgen/v1"
	"github.com/converged-computing/rainbow/pkg/graph"
	"github.com/converged-computing/rainbow/pkg/graph/algorithm"
	rspec "github.com/converged-computing/rainbow/pkg/jobspec"
	rlog "github.com/converged-computing/rainbow/pkg/logger"
	"github.com/converged-computing/rainbow/pkg/types"
	"github.com/converged-computing/rainbow/plugins/algorithms/shared"
)

// DFSForMatch WILL is a depth first search if the cluter matches
// It starts by looking at total cluster resources on the top level,
// and then traverses into those that match the first check
// THIS IS EXPERIMENTAL and likely wrong, or missing details,
// which is OK as we will only be using it for prototyping.
func (g *ClusterGraph) DFSForMatch(
	jobspec *v1.Jobspec,
	matcher algorithm.MatchAlgorithm,
) (bool, error) {

	// Get subsystem (will get dominant, this can eventually take a variable)
	subsystem := g.getSubsystem("")

	// Assume we are querying the dominant subsystem with nodes to start
	ss, ok := g.subsystem[g.dominantSubsystem]
	if !ok {
		return false, fmt.Errorf("the subsystem %s does not exist", subsystem)
	}

	// Do a quick top level count for resource types
	isMatch := true
	totals := graph.ExtractResourceSlots(jobspec)
	fmt.Println(totals)

	for resourceType, needed := range totals {
		actual, ok := ss.Metrics.ResourceCounts[resourceType]

		// We don't know. Assume we can't schedule
		if !ok {
			return false, nil
		}
		// We don't have enough resources
		if int32(actual) < needed {
			return false, nil
		}
	}
	// If it's a superficial match, search more deeply
	if isMatch {
		return g.depthFirstSearch(ss, jobspec, matcher)
	}
	return false, nil
}

// depthFirstSearch fully searches the graph finding a list of maches and a jobspec
func (g *ClusterGraph) depthFirstSearch(
	dom *Subsystem,
	jobspec *v1.Jobspec,
	matcher algorithm.MatchAlgorithm,
) (bool, error) {

	// Get resources that need scheduling from the jobspec
	// This is a map[string]Resource{} that may or may not have type slot
	resources := jobspec.GetScheduledNamedSlots()

	// If we don't have jobspec.Resources, nothing to search for
	// Return early based on top level counts
	if len(resources) == 0 {
		rlog.Debugf("  🎰️ No resources defined, top level counts satisfied so cluster is match\n")
		return true, nil
	}

	// Note that in the experimental version we have one task and thus one slot
	rlog.Debugf("  🎰️ Resources that that need to be satisfied with matcher %s\n", matcher.Name())

	// From this point on we assume we MUST satisfy the slot
	// Sanity check what we are trying to match
	for rname, rslot := range resources {
		rspec.ShowRequires(rname, &rslot)
	}

	// Look through our potential matching clusters
	rlog.Debugf("\n  🔍️ Exploring cluster %s deeper with depth first search\n", g.Name)

	// This is the root vertex of the cluster where we start search
	// It is the first Vertex that was added
	rootName := fmt.Sprintf("%s-0", dom.Name)
	root := dom.Lookup[rootName]
	vertex := dom.Vertices[root]

	// localResourceMatch checks for local edges to match
	var localResourceMatch func(vtx *types.Vertex, resourceNeeds *types.SlotResourceNeeds) bool
	localResourceMatch = func(vtx *types.Vertex, resourceNeeds *types.SlotResourceNeeds) bool {

		if len(resourceNeeds.Subsystems) == 0 || resourceNeeds.Satisfied {
			return true
		}

		// Check the vertex if it's the right type
		if vtx.Type == resourceNeeds.Type {
			for sName, edges := range vtx.Subsystems {

				// This does the check across subsystem edges
				for _, child := range edges {
					matcher.CheckSubsystemEdge(resourceNeeds, child, vtx)
				}
				// When we finish checking edges, they should all be satisfied for the subsystem
				if !resourceNeeds.IsSatisfiedFor(sName) {
					return false
				}
			}
		} else {
			for _, edge := range vtx.Edges {
				matched := localResourceMatch(edge.Vertex, resourceNeeds)
				if !matched {
					return false
				}
			}
		}
		return true
	}

	// Recursive function to recurse into slot resource "requires" and find count of matches for the slot
	// This returns a count of the matching slots under a parent level, recursing into child vertices until
	// we find the right type (and take a count) or keep exploring
	var findSlots func(vtx *types.Vertex, slot *v1.Resource, slotsFound, slotsNeeded int32) int32
	findSlots = func(vtx *types.Vertex, resource *v1.Resource, slotsFound, slotsNeeded int32) int32 {

		// Regardless of slot or not, a resource can have requirements
		resourceNeeds := shared.GetSlotResourceNeeds(resource)

		// If we already have what we need, cut out early
		if slotsFound >= slotsNeeded {
			return slotsFound
		}

		// Subsystem edges are here, separate from dominant ones (so search is smaller)
		for sName, edges := range vtx.Subsystems {
			for _, child := range edges {

				rlog.Debugf("           Found subsystem edge for %s with type %s\n", sName, child.Vertex.Type)
				// Check if the subsystem edge satisfies the needs of the slot
				// This will update the slotNeeds.Satisfied
				matcher.CheckSubsystemEdge(resourceNeeds, child, vtx)

				// Return early if minimum needs are satsified
				if resourceNeeds.Satisfied {
					rlog.Debugf("           Minimum slot needs are satisfied at %s for %s at %s, returning early.\n", vtx.Type, child.Subsystem, child.Vertex.Type)
					return slotsFound + vtx.Size
				}
			}
		}

		// Otherwise, we haven't found the right level of the graph, keep going
		for _, child := range vtx.Edges {

			// If we got what we need in the loop, cut out early
			if slotsFound >= slotsNeeded {
				return slotsFound
			}

			rlog.Debugf("      => Searching for resource type %s from parent %s->%s\n", resource.Type, child.Relation, child.Vertex.Type)

			// Only keep going if we aren't stopping here
			// This is also traversing the dominant subsystem
			if child.Relation == types.ContainsRelation {
				slotsFound += findSlots(child.Vertex, resource, slotsFound, slotsNeeded)
			}
		}

		// Stop here is true when we are at the slot -> one level below, and
		// have done the subsystem assessment on this level.
		if resourceNeeds.Satisfied {
			rlog.Debugf("         slotNeeds are satisfied, returning %d slots matched\n", slotsFound)
			return slotsFound
		}
		rlog.Debugf("         slotNeeds are not satisfied, returning 0 slots matched\n", resourceNeeds.Subsystems)
		return 0
	}

	// Traverse resource is the main function to handle traversing a cluster vertex
	// TODO this should take a vertex so we can target the child nodes
	var traverseResource func(resource v1.Resource, vtx *types.Vertex) (bool, error)
	traverseResource = func(resource v1.Resource, vtx *types.Vertex) (bool, error) {

		// Regardless of slot or not, a resource can have requirements
		resourceNeeds := shared.GetSlotResourceNeeds(&resource)

		// Check resource at this level if the types match
		if !localResourceMatch(vtx, resourceNeeds) {
			rlog.Debugf("         Resource needs for %s are not satisfied\n", resource.Type)
			return false, nil
		}

		// Replicas indicate the start of a slot
		if resource.Replicas > 0 {

			// Suggestion - slot count might be more clear in jobspec v2.
			// https://flux-framework.readthedocs.io/projects/flux-rfc/en/latest/spec_14.html
			// These are logical groups of "stuff" that need to be scheduled together
			slotsNeeded := resource.Replicas

			// Keep going until we have all the slots, or we run out of places to look
			slotsFound := int32(0)

			// This assumes that the slot value is defined in the next resource block
			// We assume the resources defined under the slot are needed for the slot
			if resource.With != nil {
				for _, subresource := range resource.With {
					slotsFound += findSlots(vtx, &subresource, slotsFound, slotsNeeded)
					rlog.Debugf("Slots found %d/%d for vertex %s\n", slotsFound, slotsNeeded, vtx.Type)
					if slotsFound >= slotsNeeded {
						return true, nil
					}
				}
			}
			// The slot is satisfied and we can continue searching resources
			return slotsFound >= slotsNeeded, nil

		}

		// Do the same for with children
		if resource.With != nil {
			for _, with := range resource.With {
				isMatch, err := traverseResource(with, vtx)
				if err != nil {
					return false, err
				}
				if isMatch {
					return true, nil
				}
			}
		}
		return false, nil
	}

	// Go through jobspec resources and determine satisfiability,
	// assessing each unit of work (resource group) separately
	for _, resource := range resources {
		isMatch, err := traverseResource(resource, vertex)
		if err != nil {
			return false, err
		}
		// Cut out early with yes if we found a match
		if isMatch {
			return true, nil
		}
	}
	// If we get here, no match
	return false, nil
}
