package types

import (
	"github.com/converged-computing/jsongraph-go/jsongraph/metadata"
)

// A Resource is a collection of attributes we load from a node
// intending to put into the graph, and associated functions
type Resource struct {
	Size int32
	Type string
	Unit string
	// The request coming in can know about the type
	Metadata metadata.Metadata
}

// A Cluster state is a key value interface.
// The algorithms are required to know what they are looking for
type ClusterState map[string]interface{}

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

// Serialize slot resource needs into a struct that is easier to parse
type SlotResourceNeeds struct {
	Satisfied  bool
	Subsystems []SubsystemNeeds
}

type SubsystemNeeds struct {
	Name       string
	Attributes map[string]bool
}
