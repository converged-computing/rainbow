package match

import (
	"strings"

	"github.com/converged-computing/rainbow/pkg/graph/algorithm"
	rlog "github.com/converged-computing/rainbow/pkg/logger"
	"github.com/converged-computing/rainbow/pkg/types"
	"github.com/converged-computing/rainbow/plugins/algorithms/equals"
	rangematch "github.com/converged-computing/rainbow/plugins/algorithms/range"
)

type MatchType struct{}

var (
	description = "match single values or ranges for subsystem job assignment"
	matcherName = "match"
)

func (s MatchType) Name() string {
	return matcherName
}

func (s MatchType) Description() string {
	return description
}

// checkSubsystemEdge evaluates a node edge in the dominant subsystem for a
// subsystem attribute. E.g., if the io subsystem provides
// Vertex (from dominant subsysetem) is only passed in for informational purposes
func (m MatchType) CheckSubsystemEdge(slotNeeds *types.SlotResourceNeeds, edge *types.Edge, vtx *types.Vertex) {

	// Return early if we are satisfied
	if slotNeeds.Satisfied {
		return
	}
	// Determine if our slot needs can be met
	// Nested for loops are not great - this will be improved with a more robust graph
	// that isn't artisinal avocado toast developed by me :)

	rlog.Debugf("Looking at edge %s->%s\n", edge.Relation, edge.Vertex.Type)

	// TODO Keep a record if all are satisfied so we stop searching
	// earlier if this is the case on subsequent calls
	for i, subsys := range slotNeeds.Subsystems {

		rlog.Debugf("      => Looking in subsystem %s\n", edge.Subsystem)

		// The subsystem has an edge defined here!
		if subsys.Name == edge.Subsystem {
			rlog.Debugf("      => Found matching subsystem %s edge\n", edge.Subsystem)

			// Yuck, this needs to be a query! Oh well.
			for k := range subsys.Attributes {
				rlog.Debugf("      => Looking at edge %s '%s' for %s that needs %s\n", edge.Subsystem, edge.Vertex.Type, subsys.Name, k)

				// We care if the attribute is marked as a range or exact match
				if strings.HasPrefix(k, "match") {
					m := equals.EqualsType{}
					m.MatchEdge(k, edge, &subsys)

				} else if strings.HasPrefix(k, "range") {
					m := rangematch.RangeType{}
					m.MatchEdge(k, edge, &subsys)
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

// Init provides extra initialization functionality, if needed
// The in memory database can take a backup file if desired
func (s MatchType) Init(options map[string]string) error {
	// If an algorithm has options, they can be set here
	return nil
}

// Add the selection algorithm to be known to rainbow
func init() {
	algo := MatchType{}
	algorithm.Register(algo)
}
