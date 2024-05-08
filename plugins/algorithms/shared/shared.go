package shared

import (
	v1 "github.com/compspec/jobspec-go/pkg/nextgen/v1"
	rlog "github.com/converged-computing/rainbow/pkg/logger"
	"github.com/converged-computing/rainbow/pkg/types"
	"github.com/converged-computing/rainbow/plugins/algorithms/equals"
	rangematch "github.com/converged-computing/rainbow/plugins/algorithms/range"
)

// getSlotResourceNeeds converts the string values into SlotResourceNeeds
// it is an explicit match, so we expect the slot to have that exact resource
// available. This can eventually take a count, but right now is a boolean match
// and this is done intentionally to satisfy the simplest scheduler experiment
// prototype where we are more interested in features
func GetSlotResourceNeeds(resources *v1.Resource) *types.SlotResourceNeeds {
	sNeeds := map[string]map[string]bool{}

	for _, needs := range resources.Requires {

		// The name of the subsystem has to be under name
		subsystem, ok := needs["name"]
		if !ok {
			continue
		}

		// This is the meta (combined) matcher that support other matchers
		// This also assumes we can have a match and range block
		sNeeds = equals.UpdateResourceNeeds(needs, subsystem, sNeeds)
		sNeeds = rangematch.UpdateResourceNeeds(needs, subsystem, sNeeds)
	}
	// Parse into the slot resource needs
	resourceNeeds := []types.SubsystemNeeds{}
	for subsystem, sneeds := range sNeeds {
		subsystemNeeds := types.SubsystemNeeds{Name: subsystem, Attributes: sneeds}
		resourceNeeds = append(resourceNeeds, subsystemNeeds)
	}

	// If we don't have any needs, the slot is satisfied for that
	slotNeeds := &types.SlotResourceNeeds{Subsystems: resourceNeeds, Type: resources.Type}
	if len(resourceNeeds) == 0 {
		slotNeeds.Satisfied = true
	} else {
		rlog.Debugf("      => Assessing needs for slot: %v\n", slotNeeds.Subsystems)
	}
	return slotNeeds
}
