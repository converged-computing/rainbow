package memory

import (
	"fmt"
	"log"

	jgf "github.com/converged-computing/jsongraph-go/jsongraph/v2/graph"
)

// addNode (vertices) to the cluster graph
func (g *ClusterGraph) addNodes(
	name string,
	nodes *jgf.JsonGraph,
	subsystem string,
) (*Subsystem, map[string]int, error) {

	// We will return a lookup of the raw (not namespaced) vertices
	lookup := map[string]int{}

	// Fall back to dominant subsystem name
	subsystem = g.getSubsystem(subsystem)

	// Let's be pedantic - no clusters allowed without nodes or edges
	err, nNodes, nEdges := g.validateNodes(nodes)
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

	// Get the root vertex, every new subsystem starts there!
	// The root vertex is named according to the subsystem
	root, exists := ss.GetNode(subsystem)
	if !exists {
		return ss, lookup, fmt.Errorf("root node does not exist for subsystem %s, this should not happen", subsystem)
	}

	// The cluster root can only exist as one, we don't want to delete given
	// references from subsystems for now (will need another function to delete)
	_, ok = ss.Lookup[name]
	if ok {
		return ss, lookup, fmt.Errorf("cluster %s already exists, delete first", name)
	}

	// Create an empty resource counter for the cluster
	ss.Metrics.NewResource(name)

	// Now loop through the nodes and add them, keeping a temporary lookup
	lookup[subsystem] = root
	for nid, node := range nodes.Graph.Nodes {

		// Currently we are saving the type, size, and unit
		resource := NewResource(node)

		// levelName (cluster)
		lookupName := getNamespacedName(name, nid)

		// If it's the cluster, we save the named identifier for it
		// We aren't interested in other metadata here so we don't add it
		id := ss.AddNode(
			name,
			lookupName,
			resource.Type,
			resource.Size,
			resource.Unit,
			resource.Metadata,
			true,
		)
		lookup[nid] = id

		// If it's a cluster, connect to the root
		if resource.Type == subsystem {
			err := ss.AddInternalEdge(root, id, 0, containsRelation, g.dominantSubsystem)
			if err != nil {
				return ss, lookup, err
			}
		}
	}
	return ss, lookup, nil
}
