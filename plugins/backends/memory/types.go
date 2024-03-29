package memory

import (
	"github.com/converged-computing/jsongraph-go/jsongraph/metadata"
)

// A subsystem is a graph with a set of vertices that are connected by edges
// We use "vertex" instead of node to distinguish the graph vs.
// a compute note
type Subsystem struct {

	// Name of the subsystem
	Name string

	// Using a map means O(1) lookup time
	Vertices map[int]*Vertex `json:"vertices"`

	// There are a small number of vertices we care to lookup by name
	// Put them here for now until I have a better idea :)
	Lookup map[string]int

	// Simple counter for adding the next code
	counter int

	// Subsystem level metrics
	Metrics Metrics
}

// A Resource is a collection of attributes we load from a node
// intending to put into the graph, and associated functions
type Resource struct {
	Size int32
	Type string
	Unit string
	// The request coming in can know about the type
	Metadata metadata.Metadata
}

// A vertex is defined by an identifier. We use an int
// instead of a string because it's faster. Edges are other
// vertices (and their identifiers) it's connected to.
type Vertex struct {
	Identifier int           `json:"identifier"`
	Edges      map[int]*Edge `json:"edges"`
	Size       int32         `json:"size"`
	Unit       string        `json:"unit"`
	Type       string        `json:"type"`

	// Link to another subsystem vertex
	Subsystems map[string]map[int]*Edge `json:"subsystems"`

	// Less commonly accessed (and standardized) metadaa
	Metadata metadata.Metadata
}

// An edge in the graph has a source vertex (where it's defined from)
// and a destination (the Vertex field below)
type Edge struct {
	Weight    int     `json:"weight"`
	Vertex    *Vertex `json:"vertex"`
	Relation  string  `json:"relation"`
	Subsystem string  `json:"subsystem"`
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

// Serialize slot resource needs into a struct that is easier to parse
type SlotResourceNeeds struct {
	Satisfied  bool
	Subsystems []SubsystemNeeds
}

type SubsystemNeeds struct {
	Name       string
	Attributes map[string]bool
}
