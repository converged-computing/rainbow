package memory

import (
	"fmt"

	v1 "github.com/compspec/jobspec-go/pkg/jobspec/experimental"
	"github.com/converged-computing/rainbow/pkg/graph/algorithm"
	"github.com/converged-computing/rainbow/pkg/types"
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

	// Get the summary metrics for the subsystem
	fmt.Println(ss.Metrics.ResourceCounts)

	isMatch := true
	for resourceType, needed := range totals {

		// TODO this should be part of a subsystem spec to ignore
		if resourceType == "slot" {
			continue
		}

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

	// Note that in the experimental version we have one task and thus one slot
	if !g.quiet {
		fmt.Printf("  🎰️ Slots that need to be satisfied with matcher %s\n", matcher.Name())
	}
	slots := map[string]*v1.Task{}

	// If a slot isn't defined for the task, assume the slot is at the top level
	topLevel := false
	if jobspec.Task.Slot == "" {
		topLevel = true
		jobspec.Task.Slot = "root"
	}
	slots[jobspec.Task.Slot] = &jobspec.Task

	// If we don't have jobspec.Task.Resources, no slot to search for.
	// Return early based on top level counts
	if len(jobspec.Task.Resources) == 0 {
		if !g.quiet {
			fmt.Printf("  🎰️ No resources defined, top level counts satisfied so cluster is match\n")
		}
		return true, nil
	}

	// From this point on we assume we MUST satisfy the slot
	// Sanity check what we are trying to match
	for rname, rslot := range jobspec.Task.Resources {
		fmt.Printf("     %s: %s\n", rname, rslot)
	}

	// Look through our potential matching clusters
	if !g.quiet {
		fmt.Printf("\n  🔍️ Exploring cluster %s deeper with depth first search\n", g.Name)
	}
	// This is the root vertex of the cluster "cluster" we start with it
	// We can store this instead, but for now we can assume the index 0
	// is the root, as it is the first one made / added
	rootName := fmt.Sprintf("%s-0", dom.Name)
	root := dom.Lookup[rootName]
	vertex := dom.Vertices[root]

	// Recursive function to recurse into slot resource and find count
	// of matches for the slot. This returns a count of the matching
	// slots under a parent level, recursing into child vertices until
	// we find the right type (and take a count) or keep exploring
	var findSlots func(vtx *types.Vertex, slot *v1.Resource, slotNeeds *types.SlotResourceNeeds, slotsFound int32) int32
	findSlots = func(vtx *types.Vertex, resource *v1.Resource, slotNeeds *types.SlotResourceNeeds, slotsFound int32) int32 {

		// This is just for debugging
		lookingFor := ""
		for _, item := range slotNeeds.Subsystems {
			lookingFor += fmt.Sprintf("%s:%v", item.Name, item.Attributes)
		}

		// Subsystem edges are here, separate from dominant ones (so search is smaller)
		for sName, edges := range vtx.Subsystems {
			// fmt.Printf("      => Searching for %s and resource type %s in subsystem %v with %d subsystem edges\n", lookingFor, resource.Type, sName, len(edges))

			for _, child := range edges {
				if !g.quiet {
					fmt.Printf("         Found subsystem edge %s with type %s\n", sName, child.Vertex.Type)
				}
				// Check if the subsystem edge satisfies the needs of the slot
				// This will update the slotNeeds.Satisfied
				matcher.CheckSubsystemEdge(slotNeeds, child, vtx)

				// Return early if minimum needs are satsified
				if slotNeeds.Satisfied {
					if !g.quiet {
						fmt.Printf("         Minimum slot needs are satisfied at %s for %s at %s, returning early.\n", vtx.Type, child.Subsystem, child.Vertex.Type)
					}
					return slotsFound + vtx.Size
				}
			}
		}

		// Otherwise, we haven't found the right level of the graph, keep going
		for _, child := range vtx.Edges {
			//fmt.Printf("      => Searching for %s and resource type %s %s->%s\n", lookingFor, resource.Type, child.Relation, child.Vertex.Type)

			// Only keep going if we aren't stopping here
			// This is also traversing the dominant subsystem
			if child.Relation == containsRelation {
				slotsFound += findSlots(child.Vertex, resource, slotNeeds, slotsFound)
			}
		}

		// Stop here is true when we are at the slot -> one level below, and
		// have done the subsystem assessment on this level.
		if slotNeeds.Satisfied {
			//fmt.Printf("         slotNeeds are satsified for %v, returning %d slots matched\n", slotNeeds.Subsystems, slotsFound)
			return slotsFound
		}
		//fmt.Printf("         slotNeeds are not satsified for %v, returning 0 slots matched\n", slotNeeds.Subsystems)
		return 0
	}

	// Traverse resource is the main function to handle traversing a cluster vertex
	var traverseResource func(resource *v1.Resource) (bool, error)
	traverseResource = func(resource *v1.Resource) (bool, error) {

		// Since we know we are looking for matching slots, we only care to check if we find one!
		if resource.Type == "slot" {

			slot, ok := slots[resource.Label]
			if !ok {
				return false, fmt.Errorf("cannot find slot %s in jobspec", resource.Label)
			}
			// Create a simple means to determine if a subsystem is matched
			// This will eventually be more complex, but right now we are just
			// matching labels, because that's all we need for the early
			// scheduling experiments. This can eventually be a setting, but right
			// now is a single algorithm (function) since there is only one.
			slotResourceNeeds := matcher.GetSlotResourceNeeds(slot)

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
			// We assume the resources defined under the slot are needed for the slot
			if resource.With != nil {
				for _, subresource := range resource.With {
					slotsFound += findSlots(vertex, &subresource, slotResourceNeeds, slotsFound)
					if !g.quiet {
						fmt.Printf("Slots found %d/%d for vertex %s\n", slotsFound, slotsNeeded, vertex.Type)
					}
				}
			}
			// The slot is satisfied and we can continue searching resources
			return slotsFound >= slotsNeeded, nil

		} else {

			// Do the same for with children
			if resource.With != nil {
				for _, with := range resource.With {
					return traverseResource(&with)
				}
			}
		}
		return false, nil
	}

	// If we need to place the slot on the top level, do it here
	if topLevel {
		newSlot := v1.Resource{
			Type:  "slot",
			Label: "root",
			Count: 1,
			With:  jobspec.Resources,
		}
		jobspec.Resources = []v1.Resource{newSlot}
	}

	// Go through jobspec resources and determine satisfiability
	// This currently treats each item under resources separately
	// as opposed to one unit of work, and I'm not sure if that is
	// right. I haven't seen jobspecs in the wild with two entries
	// under resources.
	for _, resource := range jobspec.Resources {
		isMatch, err := traverseResource(&resource)
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
