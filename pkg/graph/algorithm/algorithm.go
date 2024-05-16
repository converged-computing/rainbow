package algorithm

// An algorithm is used to match a subsystem to a slot

import (
	"fmt"
	"log"

	"github.com/converged-computing/rainbow/pkg/types"
)

// Lookup of Algorthms
var (
	MatchAlgorithms map[string]MatchAlgorithm
)

// A SelectionAlgorithm is used by the rainbow scheduler to make
// a final decision about assigning work to a group of clusters.
type MatchAlgorithm interface {
	Name() string
	Description() string
	Init(map[string]string) error

	// A MatchAlgorithm needs to take a slot and determine if it matches
	CheckSubsystemEdge(slotNeeds *types.MatchAlgorithmNeeds, edge *types.Edge)

	// Graph backends that support cypher need a cypher query for the algorithm
	GenerateCypher(matchNeeds *types.MatchAlgorithmNeeds) string
}

// List returns known algorithms
func List() map[string]MatchAlgorithm {
	return MatchAlgorithms
}

// Register a new backend by name
func Register(algorithm MatchAlgorithm) {
	if MatchAlgorithms == nil {
		MatchAlgorithms = make(map[string]MatchAlgorithm)
	}
	MatchAlgorithms[algorithm.Name()] = algorithm
}

// Get a backend by name
func Get(name string) (MatchAlgorithm, error) {
	for algoName, entry := range MatchAlgorithms {
		if algoName == name {
			return entry, nil
		}
	}
	return nil, fmt.Errorf("did not find algorithm named %s", name)
}

// GetOrFail ensures we can find the entry
func GetOrFail(name string) MatchAlgorithm {
	algorithm, err := Get(name)
	if err != nil {
		log.Fatalf("Failed to get algorithm: %v", err)
	}
	return algorithm
}
