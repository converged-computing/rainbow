package memory

import (
	"encoding/gob"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	jgf "github.com/converged-computing/jsongraph-go/jsongraph/v2/graph"
	"github.com/converged-computing/rainbow/backends/memory/service"
	"github.com/converged-computing/rainbow/pkg/graph"
	"github.com/converged-computing/rainbow/pkg/utils"
)

// A ClusterGraph holds one or more subsystems
// TODO add support for >1 subsystem, start with dominant
type ClusterGraph struct {
	subsystem  map[string]*Subsystem
	lock       sync.RWMutex
	backupFile string

	// TODO: cluster level metrics?

	// The dominant subsystem is a lookup in the subsystem map
	// It defaults to nodes (node resources)
	dominantSubsystem string
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

// NewClusterGraph creates a new cluster graph with a dominant subsystem
// TODO we will want a function that can add a new subsystem
func NewClusterGraph() *ClusterGraph {

	// For now, the dominant subsystem is hard coded to be nodes (resources)
	dominant := "nodes"
	subsystem := NewSubsystem()
	subsystems := map[string]*Subsystem{dominant: subsystem}

	// TODO options / algorithms can come from config
	g := &ClusterGraph{
		subsystem:         subsystems,
		dominantSubsystem: dominant,
	}
	// Listen for syscalls to exit
	g.awaitExit()

	// Load backup file, if exists
	g.LoadBackup()
	return g
}

// await listens for syscalls and exits when they happen
func (g *ClusterGraph) awaitExit() {
	var stopper = make(chan os.Signal, 1)
	signal.Notify(stopper, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		sig := <-stopper
		fmt.Printf("memory graph database caught signal: %+v\n", sig)
		err := g.Close()
		if err != nil {
			log.Println(err.Error())
		}
		os.Exit(0)
	}()
}

// Close the database and save to backup file
func (g *ClusterGraph) Close() error {

	// No backup file, nothing to save
	if g.backupFile == "" {
		return nil
	}
	fp, err := os.Create(g.backupFile)
	if err != nil {
		return err
	}

	// a "gob" is "binary values exchanged between an encode->decoder"
	encoder := gob.NewEncoder(fp)
	defer func() {
		recovered := recover()
		if recovered != nil {
			err = errors.New("error registering item types with Gob library")
		}
	}()
	g.lock.Lock()
	defer g.lock.Unlock()

	err = encoder.Encode(&g.subsystem)
	if err != nil {
		fp.Close()
		return err
	}
	return fp.Close()
}

// GetMetrics for a named subsystem, defaulting to dominant
func (g *ClusterGraph) GetMetrics(subsystem string) Metrics {
	subsystem = g.getSubsystem(subsystem)
	ss := g.subsystem[subsystem]
	return ss.Metrics
}

// Register cluster should:
// 1. Load in json graph of nodes from string
// 2. Add nodes to the graph, also keep top level metrics?
// 3. Return corresponding response
func (g *ClusterGraph) RegisterCluster(
	name string,
	payload string,
	subsystem string,
) (*service.Response, error) {

	// Prepare a response
	response := service.Response{}

	// Load payload into jgf
	nodes, err := graph.ReadNodeJsonGraphString(payload)
	if err != nil {
		return nil, errors.New("cluster nodes are invalid")
	}

	// Load jgf into graph for that subsystem!
	err = g.LoadClusterNodes(name, &nodes, subsystem)

	// do something with g.subsystem
	return &response, err
}

func (g *ClusterGraph) LoadClusterNodes(
	name string,
	nodes *jgf.JsonGraph,
	subsystem string,
) error {

	// Fall back to dominant subsystem name
	subsystem = g.getSubsystem(subsystem)

	// Let's be pedantic - no clusters allowed without nodes or edges
	nNodes := len(nodes.Graph.Nodes)
	nEdges := len(nodes.Graph.Edges)
	if nEdges == 0 || nNodes == 0 {
		return fmt.Errorf("cluster must have at least one edge and node")
	}

	// Grab the current subsystem - it must exist
	ss, ok := g.subsystem[subsystem]
	if !ok {
		return fmt.Errorf("subsystem %s does not exist. Ensure it is created first", subsystem)
	}

	g.lock.Lock()
	defer g.lock.Unlock()
	log.Printf("Preparing to load %d nodes and %d edges\n", nNodes, nEdges)

	// Get the root vertex, every new subsystem starts there!
	root, exists := ss.GetNode("root")
	if !exists {
		return fmt.Errorf("root node does not exist for subsystem %s, this should not happen", subsystem)
	}

	// The cluster root can only exist as one, and needs to be deleted if it does.
	_, ok = ss.lookup[name]
	if ok {
		log.Printf("cluster %s already exists, cleaning up\n", name)
		delete(ss.lookup, name)
	}

	// Create an empty resource counter for the cluster
	ss.Metrics.NewResource(name)

	// Add a cluster root to it, and connect to the top root. We can add metadata/weight here too
	clusterRoot := ss.AddNode("", name, "cluster", 1, "")
	err := ss.AddEdge(root, clusterRoot, 0, "")
	if err != nil {
		return err
	}

	// Now loop through the nodes and add them, keeping a temporary lookup
	lookup := map[string]int{"root": root, name: clusterRoot}

	// This is pretty dumb because we don't add metadata yet, oh well
	// we will!
	for nid, node := range nodes.Graph.Nodes {

		// Currently we are saving the type, size, and unit
		resource := NewResource(node)

		// levelName (cluster)
		// name for lookup/cache (if we want to keep it there)
		// resource type, size, and unit
		id := ss.AddNode(name, "", resource.Type, resource.Size, resource.Unit)
		lookup[nid] = id
	}

	// Now add edges
	for _, edge := range nodes.Graph.Edges {

		// Get the nodes in the lookup
		src, ok := lookup[edge.Source]
		if !ok {
			return fmt.Errorf("source %s is defined as an edge, but missing as node in graph", edge.Label)
		}
		dest, ok := lookup[edge.Target]
		if !ok {
			return fmt.Errorf("destination %s is defined as an edge, but missing as node in graph", edge.Label)
		}
		err := ss.AddEdge(src, dest, 0, edge.Relation)
		if err != nil {
			return err
		}
	}
	log.Printf("We have made an in memory graph (subsystem %s) with %d vertices!", subsystem, ss.CountVertices())
	g.subsystem[subsystem] = ss

	// Show metrics
	ss.Metrics.Show()
	return nil
}

// LoadBackup loads the saved database from a backup
func (g *ClusterGraph) LoadBackup() error {

	// No backup file, no need to save
	if g.backupFile == "" {
		return nil

	}
	exists, err := utils.PathExists(g.backupFile)
	if !exists || err != nil {
		return err
	}

	fp, err := os.Open(g.backupFile)
	if err != nil {
		return err
	}
	defer fp.Close()

	// Load the subsystem from the filesystem gob
	dec := gob.NewDecoder(fp)
	items := g.subsystem
	err = dec.Decode(&items)
	if err == nil {
		g.lock.Lock()
		defer g.lock.Unlock()
		g.subsystem = items
	}
	return err
}
