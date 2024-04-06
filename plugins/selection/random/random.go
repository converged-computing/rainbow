package random

import (
	"math/rand"

	"github.com/converged-computing/rainbow/pkg/graph/selection"
	"github.com/converged-computing/rainbow/pkg/types"
)

// Random selection of a cluster
// It doesn't get simpler than this!

type RandomSelection struct{}

var (
	description  = "random selection of cluster for job assignment"
	selectorName = "random"
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
	jobspec string,
) (string, error) {
	if len(contenders) == 0 {
		return "", nil
	}

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
