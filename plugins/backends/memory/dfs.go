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

	for _, slotCount := range totals {
		actual, ok := ss.Metrics.ResourceCounts[slotCount.Name]
		needed := slotCount.Count

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
	rlog.Debugf("  🎰️ Resources that need to be satisfied with matcher %s\n", matcher.Name())

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
	var localResourceMatch = func(vtx *types.Vertex, resourceNeeds *types.ResourceNeeds) bool {

		// Cut out early if the subsystem needs are satisfied
		// This is above a slot, so we aren't counting containment types yet
		if resourceNeeds.AreSubsystemsSatisfied() {
			return true
		}

		// Now we want to check edges, looking for subsystem requirements
		// Is the containment vertex the right type for the resource needs?
		// For this parsing of resource needs, there is only one type
		if vtx.Type == resourceNeeds.Type {

			// If yes, look for edges to subsystem graphs
			for _, edges := range vtx.Subsystems {

				// This does the check across subsystem edges
				for _, child := range edges {

					// This is a bad design, get back updated structure
					// and explicitly put back and ensure checked for satisfy
					matchNeeds := resourceNeeds.Subsystems[vtx.Type]
					matcher.CheckSubsystemEdge(&matchNeeds, child)
					resourceNeeds.Subsystems[vtx.Type] = matchNeeds

					// As soon as all subsystems are satisfied, we can return true
					if resourceNeeds.AreSubsystemsSatisfied() {
						rlog.Debugf("           All subsystem requirements are satisfied for resource type %s\n", vtx.Type)
						return true
					}
				}
			}
		}
		// We only parse one level here, because the recursion for findSlot will check the next
		// one against the same resources (at a different vertex)
		return true
	}

	// traverseVertex looks over the top of a vertex where we can find slots
	// and finds them, also checking for subsystem requirements
	var traverseVertex func(vtx *types.Vertex, needs *types.ResourceNeeds) bool
	traverseVertex = func(vtx *types.Vertex, needs *types.ResourceNeeds) bool {

		// Note that this summary function is likely slow (to print) but very useful
		rlog.Debugf("           Checking needs at vertex %-6s %s\n", vtx.Type, needs.SummarizeRemaining())

		// First check top level of vertex if there are subsystem needs
		// This vertex checks for resources and subsystems at this level
		// If subsystem needs aren't met, it returns false. If needs are
		// met OR the vertex type isn't relevant we return true and continue
		// If OK, we continue. If not, we stop.
		if !shared.CheckVertex(needs, vtx) {
			return false
		}

		// Now check if the resources AND subsystem needs are all satisfied
		if needs.AllSatisfied() {
			needs.Found += 1
			needs.Reset()
			if needs.Satisfied() {
				return true
			}
		}

		// If we get here, the needs aren't all satisfied, so keep recursing into the children
		for _, edge := range vtx.Edges {

			// Only interested in containment subsystem node
			if edge.Subsystem != types.DefaultDominantSubsystem {
				continue
			}
			if traverseVertex(edge.Vertex, needs) {
				return true
			}

		}
		return false
	}

	// findSlot is the main function to handle finding the top level slot
	// 1. Traverse the graph until we find the right level of the slot. As we
	//    traverse, we check requires and exit early if our parent resources
	//    don't meet requirements.
	var findSlot func(resource v1.Resource, vtx *types.Vertex) (bool, error)
	findSlot = func(resource v1.Resource, vtx *types.Vertex) (bool, error) {

		// We assume at this point we haven't yet found the slot. This function
		// summarizes resource needs by type, so we can check the type and then
		// the edges it has (and cut out early if not a match)
		resourceNeeds := shared.GetResourceNeeds(&resource)

		// Check resource at this level if the types match
		if !localResourceMatch(vtx, resourceNeeds) {
			rlog.Debugf("         Resource needs for %s are not satisfied\n", resource.Type)
			return false, nil
		}

		// If replicas are 0, we need to find the slot in a child
		// We've already checked subsystem requirements for the "non slot" vertices
		// above.
		if resource.Replicas == 0 {
			if resource.With != nil {
				for _, with := range resource.With {
					isMatch, err := findSlot(with, vtx)
					if err != nil {
						return false, err
					}
					if isMatch {
						return true, nil
					}
				}
			}
			// If we don't find matches, or if there aren't resource "with"
			// to find a slot, we don't have anything to match to
			return false, nil
		}

		rlog.Debugf("         Scheduling slot found at level %s\n", resource.Type)
		// If we get here, we found a slot. Prepare to search it.
		// This will keep track of:
		// 1. subsystem needs for resource types within a slot
		// 2. counts for resource types
		// 3. slots satisfied vs. needed
		// Unlike "GetResourceNeeds" above, this recurses the entire resource
		slotNeeds := shared.GetSlotNeeds(&resource)

		// A slot is a logical groups of "stuff" that needs to be scheduled together
		slotNeeds.Needed = resource.Replicas

		// This assumes that the slot value is defined in the next resource block
		// We assume the resources defined under the slot are needed for the slot
		for _, edge := range vtx.Edges {
			// If the slots and counts are satisfied on a traversal, return early.
			if traverseVertex(edge.Vertex, slotNeeds) {
				rlog.Debugf("         Slot needs fully satisfied on traversal of %s\n", edge.Vertex.Type)
				return true, nil
			}
		}

		// If we never get to state of all satisfied on last vertex, we are not satisfied
		return false, nil
	}

	// Go through jobspec resources and determine satisfiability,
	// We do this by first looking for the slot defined by the resource,
	// and along the way checking for requirements. Once we find a slot,
	// we traverse it and dive in to count the number of satisfied units
	// below it.
	for _, resource := range resources {

		// This always starts at the top level of the cluster (vertex is the root)
		isMatch, err := findSlot(resource, vertex)
		if err != nil {
			return false, err
		}
		// Cut out early if one resource group cannot be matched
		if !isMatch {
			return false, nil
		}
	}
	// If we get here, all groups have matched
	return true, nil
}
