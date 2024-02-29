package memory

// A subsystem is a graph with a set of vertices that are connected by edges
// We use "vertex" instead of node to distinguish the graph vs.
// a compute note
type Subsystem struct {

	// Using a map means O(1) lookup time
	Vertices map[int]*Vertex `json:"vertices"`

	// There are a small number of vertices we care to lookup by name
	// Put them here for now until I have a better idea :)
	lookup map[string]int

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
}

// An edge in the graph has a source vertex (where it's defined from)
// and a destination (the Vertex field below)
type Edge struct {
	Weight   int     `json:"weight"`
	Vertex   *Vertex `json:"vertex"`
	Relation string  `json:"relation"`
}

// Metrics keeps track of counts of things
type Metrics struct {
	// This is across all subsystems
	Vertices int   `json:"vertices"`
	Writes   int64 `json:"writes"`
	Reads    int64 `json:"reads"`

	// Resource specific metrics
	ResourceSummary map[string]Summary
}

// Resource summary to hold counts for each type
// We assemble this as we create a new graph
type Summary struct {
	Name   string
	Counts map[string]int64
}
