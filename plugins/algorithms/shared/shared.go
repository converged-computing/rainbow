package shared

import (
	v1 "github.com/compspec/jobspec-go/pkg/nextgen/v1"
	rlog "github.com/converged-computing/rainbow/pkg/logger"
	"github.com/converged-computing/rainbow/pkg/types"
	"github.com/converged-computing/rainbow/plugins/algorithms/match"
)

// getSlotNeeds converts the string values into SlotNeeds
// unlike ResourceNeeds, for a slot we also have a counter for
// different types, and the total number we have found so far.
func GetSlotNeeds(resources *v1.Resource) *types.ResourceNeeds {

	// type -> subsystem -> attribute -> boolean yes/no
	matchNeeds := map[string]types.MatchAlgorithmNeeds{}

	// type -> count
	// Assuming types not present on more than one level
	resourceNeeds := map[string]int32{}

	// Recursive to populate match needs based on type
	// This assumes that we see each type once in a resource specs
	// This may not be true for advanced cases
	var traverseResources func(resource *v1.Resource)
	traverseResources = func(resource *v1.Resource) {

		// We assume each resource block is unique in type
		// This is a lookup of needs for the type by subsystem
		typeNeeds := types.MatchAlgorithmNeeds{}

		// These are requirements for the subsystem
		for _, needs := range resource.Requires {

			// The name of the subsystem has to be under name
			subsystem, ok := needs["name"]
			if !ok {
				continue
			}
			subsystemNeeds := match.GetResourceNeeds(needs)
			if len(subsystemNeeds) > 0 {
				typeNeeds[subsystem] = subsystemNeeds
			}
		}

		if len(typeNeeds) > 0 {
			matchNeeds[resource.Type] = typeNeeds
		}
		// These are containment counts
		// This assumes no slots below slots, because this would be 0
		// in favor of replicas, which would imply 0 are needed which
		// is likely not the case.
		if resource.Count > 0 {
			resourceNeeds[resource.Type] = resource.Count
		}

		// Parse the reset
		if resource.With != nil {
			for _, subresource := range resource.With {
				traverseResources(&subresource)
			}
		}
	}
	traverseResources(resources)

	// Create new slot needs with subsystems and backup copy
	// for cache/restore when we call refresh after finding a slot
	slotNeeds := &types.ResourceNeeds{
		Subsystems:         matchNeeds,
		SubsystemsOriginal: matchNeeds,
		Resources:          resourceNeeds,
		ResourcesOriginal:  resourceNeeds,
	}

	// Do a first check to see if they are satisfied
	slotNeeds.AllSatisfied()
	return slotNeeds
}

// CheckVertex ensures that subsystem needs are satisfied (and updates them)
// and if so, includes the resource type in the count
func CheckVertex(
	slotNeeds *types.ResourceNeeds,
	vtx *types.Vertex,
) bool {

	resourceNeeds := (*slotNeeds)

	// Cut out early if the vertex type isn't in our needs
	count, ok := slotNeeds.Resources[vtx.Type]
	typeNeeds, ok2 := resourceNeeds.Subsystems[vtx.Type]

	if !ok && !ok2 {
		rlog.Debugf("             Cutting out early, vertex type %s not in needs\n", vtx.Type)
		return true
	}

	// First step is to check the vertex edges for subsystem matches
	for _, edges := range vtx.Subsystems {

		for _, edge := range edges {
			subsystemNeeds := typeNeeds[edge.Subsystem]
			for attribute, isSatisfied := range match.CheckSubsystemNeeds(subsystemNeeds, edge) {
				resourceNeeds.Subsystems[vtx.Type][edge.Subsystem][attribute] = isSatisfied
				if !isSatisfied {
					return false
				} else {
					rlog.Debugf("             Resource need for %s %s satisfied with edge %s\n", vtx.Type, attribute, edge.Vertex.Type)
				}
			}
		}
	}
	// If we get here, the vertex has the subsystem features we want
	// update the counts of resources
	count -= vtx.Size
	resourceNeeds.Resources[vtx.Type] = count
	slotNeeds = &resourceNeeds
	return true
}

// getResourceNeeds flattens a resource requirement into names
// This is intended to just check subsystem metadata for one resource
// type (e.g., node) before we have dived into a slot
func GetResourceNeeds(resources *v1.Resource) *types.ResourceNeeds {
	matchNeeds := types.MatchAlgorithmNeeds{}

	for _, needs := range resources.Requires {

		// The name of the subsystem has to be under name
		subsystem, ok := needs["name"]
		if !ok {
			continue
		}
		matchNeeds[subsystem] = match.GetResourceNeeds(needs)
	}

	// Since the type is relevant here, organize the matchNeeds by the one type
	needs := map[string]types.MatchAlgorithmNeeds{resources.Type: matchNeeds}

	// If we don't have any needs, the slot is satisfied for that
	slotNeeds := &types.ResourceNeeds{Subsystems: needs, Type: resources.Type}
	slotNeeds.AreResourcesSatisfied()
	return slotNeeds
}
