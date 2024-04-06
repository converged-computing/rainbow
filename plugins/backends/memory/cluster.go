package memory

import (
	"fmt"
	"log"
	"sync"

	jgf "github.com/converged-computing/jsongraph-go/jsongraph/v2/graph"
	rlog "github.com/converged-computing/rainbow/pkg/logger"
	"github.com/converged-computing/rainbow/pkg/types"
)

var (
	// cluster == containment == nodes
	defaultDominantSubsystem = "cluster"
	containsRelation         = "contains"
)

// A ClusterGraph holds a single graph with one or more subsystems
type ClusterGraph struct {
	subsystem map[string]*Subsystem
	lock      sync.RWMutex
	State     map[string]interface{}

	// Courtesy holder for name
	Name string

	// The dominant subsystem is a lookup in the subsystem map
	// It defaults to nodes (node resources)
	dominantSubsystem string
}

// GetState of the cluster
// We could expose this as a public variable, but I'm leaving
// like this in case we want to do additional processing
// (for example, maybe some attributes are private)
func (c *ClusterGraph) GetState() types.ClusterState {
	return c.State
}

// Dominant subsystem gets the dominant subsystem
func (c *ClusterGraph) DominantSubsystem() *Subsystem {
	return c.subsystem[c.dominantSubsystem]
}

// getSubsystem returns a named subsystem, or falls back to the default
func (c *ClusterGraph) getSubsystem(subsystem string) string {
	if subsystem == "" {
		subsystem = c.dominantSubsystem
	}
	return subsystem
}

func (g *ClusterGraph) LoadClusterNodes(
	nodes *jgf.JsonGraph,
	subsystem string,
) error {

	// Create the new subsystem for it, and add nods
	subsystem = g.getSubsystem(subsystem)
	ss, lookup, err := g.addNodes(nodes, subsystem)
	if err != nil {
		return err
	}

	// Now add edges
	for _, edge := range nodes.Graph.Edges {

		// We only care about contains for now, recursive lets us just return
		// This reduces redundancy of edges. We will need the double linking
		// for subsystems
		if edge.Relation != containsRelation {
			continue
		}

		// Get the nodes in the lookup
		src, ok := lookup[edge.Source]
		if !ok {
			return fmt.Errorf("source %s is defined as an edge, but missing as node in graph", edge.Label)
		}
		dest, ok := lookup[edge.Target]
		if !ok {
			return fmt.Errorf("destination %s is defined as an edge, but missing as node in graph", edge.Label)
		}
		rlog.Debugf("Adding edge from %s -%s-> %s\n", ss.Vertices[src].Type, edge.Relation, ss.Vertices[dest].Type)
		err := ss.AddInternalEdge(src, dest, 0, edge.Relation, subsystem)
		if err != nil {
			return err
		}
	}
	log.Printf("We have made an in memory graph (subsystem %s) with %d vertices!", subsystem, ss.CountVertices())

	// Show metrics
	ss.Metrics.Show()
	return nil
}

// validateNodes ensures that we have at least one node and edge
func (c *ClusterGraph) validateNodes(nodes *jgf.JsonGraph) (error, int, int) {
	var err error
	nNodes := len(nodes.Graph.Nodes)
	nEdges := len(nodes.Graph.Edges)
	if nEdges == 0 || nNodes == 0 {
		err = fmt.Errorf("subsystem cluster must have at least one edge and node")
	}
	return err, nNodes, nEdges
}

// NewClusterGraph creates a new cluster graph with a dominant subsystem
// We assume the dominant is hard coded to be containment
func NewClusterGraph(name string, domSubsystem string) *ClusterGraph {

	// If not defined, set the dominant subsystem
	if domSubsystem == "" {
		domSubsystem = defaultDominantSubsystem
	}
	// For now, the dominant subsystem is hard coded to be nodes (resources)
	subsystem := NewSubsystem(domSubsystem)
	subsystems := map[string]*Subsystem{defaultDominantSubsystem: subsystem}

	// TODO options / algorithms can come from config
	g := &ClusterGraph{
		Name:              name,
		subsystem:         subsystems,
		dominantSubsystem: defaultDominantSubsystem,
		State:             types.ClusterState{},
	}
	return g
}

// GetMetrics for a named subsystem, defaulting to dominant
func (g *ClusterGraph) GetMetrics(subsystem string) Metrics {
	subsystem = g.getSubsystem(subsystem)
	ss := g.subsystem[subsystem]
	return ss.Metrics
}

// LoadSubsystemNodes into the cluster
func (g *ClusterGraph) LoadSubsystemNodes(
	nodes *jgf.JsonGraph,
	subsystem string,
) error {

	// Get the dominant subsystem for the cluster
	dom := g.DominantSubsystem()

	// Does the subsystem exist? One unique subsytem (by name) per cluster
	_, ok := g.subsystem[subsystem]
	if ok {
		return fmt.Errorf("subsystem %s already exists for cluster %s", subsystem, g.Name)
	}
	ss := NewSubsystem(subsystem)
	g.subsystem[subsystem] = ss

	ss, lookup, err := g.addNodes(nodes, subsystem)
	if err != nil {
		return err
	}

	// Count dominant vertices references
	count := 0

	// Now add edges
	for _, edge := range nodes.Graph.Edges {

		// We are currently just saving one direction "x contains y"
		if edge.Relation != containsRelation {
			continue
		}

		// Two cases:
		// 1. the src is in the dominant subsystem
		// 2. The src is not, and both node are defined in the graph here
		subIdx1, ok1 := lookup[edge.Source]
		subIdx2, ok2 := lookup[edge.Target]

		// Case 1: both are in the subsystem graph
		if ok1 && ok2 {
			// This says "subsystem resource in node"
			fmt.Printf("Adding internal edge for %s to %s\n", edge.Source, edge.Target)
			ss.AddInternalEdge(subIdx1, subIdx2, 0, edge.Relation, subsystem)

		} else {

			// We need the namespaced name for the dom lookup
			lookupName := getNamespacedName(dom.Name, edge.Source)

			// Case 2: the src is in the dominant subsystem
			domIdx, ok := dom.Lookup[lookupName]

			fmt.Printf("Adding dominant subsystem edge for %s to %s\n", lookupName, subsystem)

			if !ok || !ok2 {
				return fmt.Errorf("edge %s->%s is not internal, and not connected to the dominant subsystem", edge.Source, edge.Target)
			}
			count += 1
			// Now add the link... the node exists in the subsystem but references a
			// different subsystem as the edge.
			// This says "dominant subsystem node conatains subsystem resource"
			err := dom.AddSubsystemEdge(domIdx, ss.Vertices[subIdx2], 0, edge.Relation, subsystem)
			if err != nil {
				return err
			}
		}
	}
	log.Printf("We have made an in memory graph (subsystem %s) with %d vertices, with %d connections to the dominant!", subsystem, ss.CountVertices(), count)
	g.subsystem[subsystem] = ss

	// Show metrics
	ss.Metrics.Show()
	return nil
}
