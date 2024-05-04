package memory

import (
	"fmt"
	"log"

	"github.com/converged-computing/rainbow/pkg/graph"
	"github.com/converged-computing/rainbow/pkg/types"

	jgf "github.com/converged-computing/jsongraph-go/jsongraph/v2/graph"
)

// addNode (vertices) to the cluster graph for a subsystem
func (g *ClusterGraph) addNodes(
	nodes *jgf.JsonGraph,
	subsystem string,
) (*Subsystem, map[string]int, error) {

	// We will return a lookup of the raw (not namespaced) vertices
	lookup := map[string]int{}

	// Fall back to dominant subsystem name
	subsystem = g.getSubsystem(subsystem)

	// Let's be pedantic - no clusters allowed without nodes or edges
	nNodes, nEdges, err := graph.ValidateNodes(nodes)
	if err != nil {
		return nil, lookup, err
	}

	// Grab the current subsystem - it must exist
	ss, ok := g.subsystem[subsystem]
	if !ok {
		return nil, lookup, fmt.Errorf("subsystem %s does not exist. Ensure it is created first", subsystem)
	}

	g.lock.Lock()
	defer g.lock.Unlock()
	log.Printf("Preparing to load %d nodes and %d edges\n", nNodes, nEdges)

	// Create an empty resource counter for the subsystem
	ss.Metrics.NewResource(subsystem)

	// Now loop through the nodes and add them, keeping a temporary lookup
	//	lookup[subsystem] = root
	for nid, node := range nodes.Graph.Nodes {

		// Currently we are saving the type, size, and unit
		resource := types.NewResource(node)

		// Defining a lookup name means that we keep a direct index to the node in
		// the subsystem lookup. We do this for edges between subsystems so
		// they are always namespaced
		lookupName := graph.GetNamespacedName(subsystem, nid)

		// If it's the cluster, we save the named identifier for it
		// We aren't interested in other metadata here so we don't add it
		id := ss.AddNode(
			lookupName,
			resource.Type,
			resource.Size,
			resource.Unit,
			resource.Metadata,
			true,
		)
		lookup[nid] = id
	}
	return ss, lookup, nil
}
