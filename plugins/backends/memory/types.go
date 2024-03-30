package memory

import (
	"github.com/converged-computing/rainbow/pkg/types"
)

// A subsystem is a graph with a set of vertices that are connected by edges
// We use "vertex" instead of node to distinguish the graph vs.
// a compute note
type Subsystem struct {

	// Name of the subsystem
	Name string

	// Using a map means O(1) lookup time
	Vertices map[int]*types.Vertex `json:"vertices"`

	// There are a small number of vertices we care to lookup by name
	// Put them here for now until I have a better idea :)
	Lookup map[string]int

	// Simple counter for adding the next code
	counter int

	// Subsystem level metrics
	Metrics Metrics
}

// Metrics keeps track of counts of things
type Metrics struct {
	// This is across all subsystems
	Vertices int   `json:"vertices"`
	Writes   int64 `json:"writes"`
	Reads    int64 `json:"reads"`

	// Courtesy to print the subsystem name
	Name string `json:"name"`

	// Resource specific metrics
	ResourceCounts map[string]int64
}
