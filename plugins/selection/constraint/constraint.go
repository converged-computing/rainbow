package random

import (
	"math/rand"

	"github.com/converged-computing/rainbow/pkg/graph/selection"
	"github.com/converged-computing/rainbow/pkg/types"
)

// Constraint selection of a cluster.
// Here the algorithm takes the following approach:
// Provide a set of priority filters.
// Each priority filter is parsed from first to last, with first highest priority
// Matching clusters (based on matching algorithms) are parsed for priority
// In the case of no matches, we parse the next priority block
// We continue until there are no blocks left, and return the match OR
// a random selection of those left

type RandomSelection struct{}

var (
	description  = "selection based on prioritized constraints"
	selectorName = "constraint"
)

func (s RandomSelection) Name() string {
	return selectorName
}

func (s RandomSelection) Description() string {
	return description
}

// Select randomly chooses a cluster from the set
// This should not receive an empty list, but we check anyway
func (s RandomSelection) Select(
	contenders []string,
	states map[string]types.ClusterState,
) (string, error) {
	if len(contenders) == 0 {
		return "", nil
	}

	/*
		name: constraint
		options:
			priorities: |
			  - filter: "nodes_free > 0"
				calc: "build_cost=(cost_per_node_hour * (memory_gb_per_node * seconds_per_gb)/60/60))"
				sort_descending: build_cost
				select: random*/

	// TODO: can match / satisfy return subsystem metrics?
	// need to add jobspec attributes into the input of select
	// need to then parse the above and use https://github.com/Knetic/govaluate
	// to do the steps in the equation.
	// going back to sleep for a bit first

	// Select a random number the length of the slice
	idx := rand.Intn(len(contenders))
	return contenders[idx], nil
}

// Init provides extra initialization functionality, if needed
// The in memory database can take a backup file if desired
func (s RandomSelection) Init(options map[string]string) error {
	// If an algorithm has options, they can be set here
	return nil
}

// Add the selection algorithm to be known to rainbow
func init() {
	algo := RandomSelection{}
	selection.Register(algo)
}
