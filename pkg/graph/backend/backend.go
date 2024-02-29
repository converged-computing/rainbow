package backend

import (
	"fmt"
	"log"

	"github.com/converged-computing/jsongraph-go/jsongraph/v2/graph"
	"google.golang.org/grpc"
)

// Lookup of Backends.
var (
	Backends map[string]GraphBackend
)

// A Graph backend is an interface to hold rainbow clusters
// Each backend should be able to handle basic queries to request work.
//
//	Satisfies: find clusters where the work can be run
//
// We will add more endpoints as they make sense. For example, rainbow does
// not control the actual scheduling, so it cannot reserve nodes or update
// resources, it can at most determine if a cluster can satisfy and then
// either ask for an ETA or assign to it.
type GraphBackend interface {
	Name() string
	Description() string
	Init(map[string]string) error

	// TODO we need a jobspec for here
	Satisfies(string) error

	// Register an additional grpc server
	RegisterService(*grpc.Server) error

	// Add nodes for a newly registered cluster
	AddCluster(name string, nodes *graph.JsonGraph, subsystem string) error
}

// List returns known backends
func List() map[string]GraphBackend {
	return Backends
}

// Register a new backend by name
func Register(backend GraphBackend) {
	if Backends == nil {
		Backends = make(map[string]GraphBackend)
	}
	Backends[backend.Name()] = backend
}

// Get a backend by name
func Get(name string) (GraphBackend, error) {
	for backendName, entry := range Backends {
		if backendName == name {
			return entry, nil
		}
	}
	return nil, fmt.Errorf("did not find backend named %s", name)
}

// GetOrFail ensures we can find the entry
func GetOrFail(name string) GraphBackend {
	backend, err := Get(name)
	if err != nil {
		log.Fatalf("Failed to get backend: %v", err)
	}
	return backend
}
