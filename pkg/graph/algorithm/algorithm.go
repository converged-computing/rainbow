package algorithm

import (
	"fmt"
	"log"
)

// Lookup of Algorthms
var (
	Algorithms map[string]SelectionAlgorithm
)

// A SelectionAlgorithm is used by the rainbow scheduler to make
// a final decision about assigning work to a group of clusters.
type SelectionAlgorithm interface {
	Name() string
	Description() string
	Init(map[string]string) error

	// Take a list of contenders and select based on algorithm
	Select([]string) (string, error)
}

// List returns known backends
func List() map[string]SelectionAlgorithm {
	return Algorithms
}

// Register a new backend by name
func Register(algorithm SelectionAlgorithm) {
	if Algorithms == nil {
		Algorithms = make(map[string]SelectionAlgorithm)
	}
	Algorithms[algorithm.Name()] = algorithm
}

// Get a backend by name
func Get(name string) (SelectionAlgorithm, error) {
	for algoName, entry := range Algorithms {
		if algoName == name {
			return entry, nil
		}
	}
	return nil, fmt.Errorf("did not find algorithm named %s", name)
}

// GetOrFail ensures we can find the entry
func GetOrFail(name string) SelectionAlgorithm {
	algorithm, err := Get(name)
	if err != nil {
		log.Fatalf("Failed to get algorithm: %v", err)
	}
	return algorithm
}
