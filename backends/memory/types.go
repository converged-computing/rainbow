package memory

// A subsystem is a graph with a set of vertices that are connected by edges
// We use "vertex" instead of node to distinguish the graph vs.
// a compute note
type Subsystem struct {

	// Using a map means O(1) lookup time
	Vertices map[int]*Vertex
}

// A vertex is defined by an identifier. We use an int
// instead of a string because it's faster. Edges are other
// vertices (and their identifiers) it's connected to.
type Vertex struct {
	Identifier int
	Edges      map[int]*Edge
}

// An edge in the graph has a source vertex (where it's defined from)
// and a destination (the Vertex field below)
type Edge struct {
	Weight int
	Vertex *Vertex
}
