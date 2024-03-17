package memory

import (
	"fmt"

	v1 "github.com/compspec/jobspec-go/pkg/jobspec/experimental"
)

// DFSForMatch WILL is a depth first search for matches
// It starts by looking at total cluster resources on the top level,
// and then traverses into those that match the first check
// THIS IS EXPERIMENTAL and likely wrong, or missing details,
// which is OK as we will only be using it for prototyping.
func (s *Subsystem) DFSForMatch(jobspec *v1.Jobspec) ([]string, error) {

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

	// Compare against each cluster we know about.
	// Clusters are saved in the resource summary of the top level
	// dominant subsystem
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
func (s *Subsystem) depthFirstSearch(matches []string, jobspec *v1.Jobspec) ([]string, error) {

	// Prepare a lookup of tasks for slots
	// Note that in the experimental version we have one task
	slots := map[string]*v1.Task{}
	slots[jobspec.Task.Slot] = &jobspec.Task

	// Keep a list of final matches
	finalMatches := []string{}

	// Look through our potential matching clusters
	for _, cluster := range matches {
		fmt.Printf("\n  🔍️ Exploring cluster %s deeper with depth first search\n", cluster)

		// This is the root vertex of the cluster "cluster" we start with it
		root := s.Lookup[cluster]
		vertex := s.Vertices[root]

		// Assume this is a match to start
		isMatch := true

		// Recursive function to recurse into slot resource and find count
		// of matches for the slot. This returns a count of the matching
		// slots under a parent level, recursing into child vertices until
		// we find the right type (and take a count) or keep exploring
		var findSlots func(vtx *Vertex, slot *v1.Resource, slotNeeds *SlotResourceNeeds) int32
		findSlots = func(vtx *Vertex, resource *v1.Resource, slotNeeds *SlotResourceNeeds) int32 {

			// This assumes the resource
			// Is the current vertex what we need? If yes, assess if it can satisfy
			slotsFound := int32(0)
			if vtx.Type == resource.Type {

				// If we hit here, we technically have the right vertex type, but we
				// also need subsystems to be satisfied
				if !slotNeeds.Satisfied {
					return slotsFound
				}
				// How many full slots can we satisfy at this vertex?
				// This indicates that subsystems are also satisfied
				return vtx.Size

			} else {

				// Otherwise, we haven't found the right level of the graph, keep going
				for _, child := range vtx.Edges {

					// Check if the subsystem edge satisfies the needs of the slot
					checkSubsystemEdge(slotNeeds, child, vtx)

					// Only interested in children. That sounds weird.
					// This is also traversing the dominant subsystem
					if child.Relation == containsRelation && child.Subsystem == s.Name {
						slotsFound += findSlots(child.Vertex, resource, slotNeeds)
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
				slot := slots[resource.Label]
				slotResourceNeeds := getSlotResourceNeeds(slot)

				// Keep going until we have all the slots, or we run out of places to look
				return findSlots(vtx, resource, slotResourceNeeds)
			}

			// Wrong resource type, womp womp
			if vtx.Type != resource.Type {
				for _, child := range vtx.Edges {

					// Update our found count to include recursing all children
					if child.Relation == containsRelation {
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

				// We need to find the subsystem resources needed under a slot
				slot := slots[resource.Label]

				// Create a simple means to determine if a subsystem is matched
				// This will eventually be more complex, but right now we are just
				// matching labels, because that's all we need for the early
				// scheduling experiments. This can eventually be a setting, but right
				// now is a single algorithm (function) since there is only one.
				slotResourceNeeds := getSlotResourceNeeds(slot)

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
						slotsFound += findSlots(vertex, &subresource, slotResourceNeeds)

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
					reason := fmt.Sprintf("%d/%d of needed %s satisfied", foundMatches, resource.Count, resource.Type)
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
