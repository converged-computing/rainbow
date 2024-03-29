package memory

import (
	"fmt"

	v1 "github.com/compspec/jobspec-go/pkg/jobspec/experimental"
)

// getSlotResource needs assumes a subsystem request as follows:
/* tasks:
- command:
  - ior
    slot: default
    count:
    per_slot: 1
  resources:
    io:
    match:
    - type: shm
*/
// it is an explicit match, so we expect the slot to have that exact resource
// available. This can eventually take a count, but right now is a boolean match
// and this is done intentionally to satisfy the simplest scheduler experiment
// prototype where we are more interested in features
func getSlotResourceNeeds(slot *v1.Task) *SlotResourceNeeds {
	sNeeds := map[string]map[string]bool{}
	for subsystem, needs := range slot.Resources {

		// Needs should be interface{} --> map[string][]map[string]string{}
		// Assume if we cannot parse, don't consider
		needs, ok := needs.(map[string]interface{})
		if !ok {
			continue
		}

		// We currently support "match" which is an exact match of a term to resource
		match, ok := needs["match"]
		if !ok {
			continue
		}

		// Now "match" goes from interface{} -> []map[string]string{}
		matches, ok := match.([]interface{})
		if !ok {
			continue
		}

		// Finally, we just parse the list - these should be key value pairs to match exactly
		for _, entry := range matches {
			entry, ok := entry.(map[string]interface{})
			if !ok {
				continue
			}
			for key, value := range entry {
				value, ok := value.(string)

				// This algorithm only knows how to match based on type
				if key != "type" {
					continue
				}
				if ok {
					_, ok := sNeeds[subsystem]
					if !ok {
						sNeeds[subsystem] = map[string]bool{}
					}

					// This sets the starting state that the value is not satisfied
					sNeeds[subsystem][value] = false
				}
			}
		}
	}
	// Parse into the slot resource needs
	needs := []SubsystemNeeds{}
	for subsystem, sneeds := range sNeeds {
		subsystemNeeds := SubsystemNeeds{Name: subsystem, Attributes: sneeds}
		needs = append(needs, subsystemNeeds)
	}

	// If we don't have any needs, the slot is satisfied for that
	slotNeeds := &SlotResourceNeeds{Subsystems: needs}
	if len(needs) == 0 {
		slotNeeds.Satisfied = true
	}
	fmt.Printf("      => Assessing needs for slot: %v\n", slotNeeds)
	return slotNeeds
}

// checkSubsystemEdge evaluates a node edge in the dominant subsystem for a
// subsystem attribute. E.g., if the io subsystem provides
// Vertex (from dominant subsysetem) is only passed in for informational purposes
func checkSubsystemEdge(slotNeeds *SlotResourceNeeds, edge *Edge, vtx *Vertex) {

	// Return early if we are satisfied
	if slotNeeds.Satisfied {
		return
	}
	// Determine if our slot needs can be met
	// Nested for loops are not great - this will be improved with a more robust graph
	// that isn't artisinal avocado toast developed by me :)

	fmt.Printf("Looking at edge %s->%s\n", edge.Relation, edge.Vertex.Type)

	// TODO Keep a record if all are satisfied so we stop searching
	// earlier if this is the case on subsequent calls
	for i, subsys := range slotNeeds.Subsystems {

		fmt.Printf("      => Looking in subsystem %s\n", edge.Subsystem)

		// The subsystem has an edge defined here!
		if subsys.Name == edge.Subsystem {
			fmt.Printf("      => Found matching subsystem %s for %s\n", subsys.Name, edge.Subsystem)

			// Yuck, this needs to be a query! Oh well.
			for k := range subsys.Attributes {
				// fmt.Printf("      => Looking at edge %s '%s' for %s that needs %s\n", edge.Subsystem, edge.Vertex.Type, subsys.Name, k)

				if edge.Vertex.Type == k {
					fmt.Printf("      => Resource '%s' has edge '%s' satisfies subsystem %s %s\n", vtx.Type, edge.Vertex.Type, subsys.Name, k)
					subsys.Attributes[k] = true
				}
			}
		}
		slotNeeds.Subsystems[i] = subsys
	}

	// Try to avoid future checking if subsystem needs are addressed
	allSatisfied := true
	for _, subsys := range slotNeeds.Subsystems {
		for _, v := range subsys.Attributes {
			if !v {
				allSatisfied = false
				break
			}
		}
	}
	// This is going to provide a quick check to determine if the subsystem
	// is satisfied without needing to parse again
	slotNeeds.Satisfied = allSatisfied
}
