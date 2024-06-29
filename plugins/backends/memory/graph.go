package memory

import (
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	js "github.com/compspec/jobspec-go/pkg/nextgen/v1"
	jgf "github.com/converged-computing/jsongraph-go/jsongraph/v2/graph"
	"github.com/converged-computing/rainbow/pkg/graph"
	"github.com/converged-computing/rainbow/pkg/graph/algorithm"
	rlog "github.com/converged-computing/rainbow/pkg/logger"
	"github.com/converged-computing/rainbow/pkg/types"
	"github.com/converged-computing/rainbow/pkg/utils"
	"github.com/converged-computing/rainbow/plugins/backends/memory/service"
)

// A graph holds one or more named clusters
type Graph struct {
	Clusters   map[string]*ClusterGraph
	lock       sync.RWMutex
	backupFile string

	// The dominant subsystem for all clusters, if desired to set
	dominantSubsystem string
}

// GetStates for clusters in the graph
func (g *Graph) GetStates(names []string) (map[string]types.ClusterState, error) {
	states := map[string]types.ClusterState{}
	for _, name := range names {

		// Only error if the cluster isn't known
		cluster, ok := g.Clusters[name]
		if !ok {
			return states, fmt.Errorf("cluster %s does not exist", name)
		}
		states[name] = cluster.GetState()
	}
	return states, nil
}

// UpdateState updates the state of a known cluster in the graph
func (g *Graph) UpdateState(name string, state *types.ClusterState) error {
	cluster, ok := g.Clusters[name]
	if !ok {
		return fmt.Errorf("cluster %s does not exist", name)
	}
	// We always update old values
	for key, value := range *state {
		rlog.Debugf("Updating state %s to %v\n", key, value)
		cluster.State[key] = value
	}
	return nil
}

// NewGraph creates a structure that holds one or more graphs
func NewGraph() *Graph {

	// Set the dominant subsystem to cluster for now
	clusters := map[string]*ClusterGraph{}
	g := Graph{dominantSubsystem: types.DefaultDominantSubsystem, Clusters: clusters}

	// Listen for syscalls to exit
	g.awaitExit()

	// Load backup file, if exists
	g.LoadBackup()
	return &g
}

// getSubsystem returns a named subsystem, or falls back to the default
// Note that a default can be defined across clusters or on the level of
// one cluster
func (c *Graph) getSubsystem(subsystem string) string {
	if subsystem == "" {
		subsystem = c.dominantSubsystem
	}
	return subsystem
}

// Register cluster should:
// 1. Load in json graph of nodes from string
// 2. Add nodes to the graph, also keep top level metrics?
// 3. Return corresponding response
func (g *Graph) RegisterCluster(
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
	err = g.LoadClusterNodes(name, &nodes, g.getSubsystem(""))

	// do something with g.subsystem
	return &response, err
}

// LoadClusterNodes loads a new cluster into the graph
func (g *Graph) LoadClusterNodes(
	clusterName string,
	nodes *jgf.JsonGraph,
	subsystem string,
) error {

	// Do we already have the graph?
	_, ok := g.Clusters[clusterName]
	if ok {
		return fmt.Errorf("cluster graph %s already exists and cannot be added again", clusterName)
	}

	// Create a new ClusterGraph
	clusterG := NewClusterGraph(clusterName, subsystem)
	err := clusterG.LoadClusterNodes(nodes, subsystem)
	if err != nil {
		return err
	}
	g.Clusters[clusterName] = clusterG
	return nil
}

// DeleteCluster removes a cluster and subsystems entirely
func (g *Graph) DeleteCluster(clusterName string) error {

	// Do we already have the graph?
	_, ok := g.Clusters[clusterName]
	if !ok {
		return fmt.Errorf("cluster graph %s does not exist", clusterName)
	}
	delete(g.Clusters, clusterName)
	return nil
}

// DeleteCluster removes a cluster and subsystems entirely
func (g *Graph) DeleteSubsystem(clusterName, subsystem string) error {
	cluster, ok := g.Clusters[clusterName]
	if !ok {
		return fmt.Errorf("cluster graph %s does not exist", clusterName)
	}
	// Now get the subsystem
	_, ok = cluster.subsystem[subsystem]
	if !ok {
		return fmt.Errorf("cluster graph %s does not have subsystem %s", clusterName, subsystem)
	}
	delete(cluster.subsystem, subsystem)
	g.Clusters[clusterName] = cluster
	return nil
}

// Satisfy should:
// 1. Read in and populate the payload into a jobspec
// 2. Determine by way of a depth first search if we can satisfy
// 3. Return the names of the cluster
func (g *Graph) Satisfies(
	payload string,
	matcher algorithm.MatchAlgorithm,
) (*service.SatisfyResponse, error) {
	response := service.SatisfyResponse{}

	// Serialize back into Jobspec
	jobspec := js.Jobspec{}
	err := json.Unmarshal([]byte(payload), &jobspec)
	if err != nil {
		return &response, err
	}

	// Tell the user /logs we are looking for a match
	rlog.Debugf("\n🍇️ Satisfy request to Graph 🍇️\n")
	rlog.Debugf(" jobspec: %s\n", payload)
	matches := []string{}
	notMatches := []string{}

	// Determine if each cluster can match
	for clusterName, clusterG := range g.Clusters {
		isMatch, err := clusterG.DFSForMatch(&jobspec, matcher)

		// Return early if we hit an error
		if err != nil {
			response.Status = service.SatisfyResponse_RESULT_TYPE_ERROR
			return &response, err
		}
		if isMatch {
			matches = append(matches, clusterName)
		} else {
			notMatches = append(notMatches, clusterName)
			// fmt.Printf("  match: 🎯️ cluster %s does not have sufficient resources and is NOT a match\n", clusterName)
		}

	}
	if len(matches) == 0 {
		fmt.Println("  match: 😥️ no clusters could satisfy this request. We are sad")
	} else {
		fmt.Printf("  match: ✅️ there are %d matches with sufficient resources\n", len(matches))
		if len(notMatches) > 0 {
			fmt.Printf("         🎯️ there are %d clusters that do not match\n", len(notMatches))
		}
	}
	// Add the matches to the response
	response.Clusters = matches
	response.TotalClusters = int32(len(g.Clusters))
	response.TotalMatches = int32(len(matches))
	response.TotalMismatches = int32(len(notMatches))
	response.Status = service.SatisfyResponse_RESULT_TYPE_SUCCESS
	return &response, nil
}

// await listens for syscalls and exits when they happen
func (g *Graph) awaitExit() {
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
func (g *Graph) Close() error {

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

	err = encoder.Encode(&g.Clusters)
	if err != nil {
		fp.Close()
		return err
	}
	return fp.Close()
}

// LoadBackup loads the saved database from a backup
func (g *Graph) LoadBackup() error {

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
	items := g.Clusters
	err = dec.Decode(&items)
	if err == nil {
		g.lock.Lock()
		defer g.lock.Unlock()
		g.Clusters = items
	}
	return err
}

// LoadSubsystemNodes into the graph
// For addition, we can have a two way pointer from the subsystem node TO
// the dominant node and then back:
// The pointer TO the dominant subsystem let's us find it to delete the opposing one
// The other one is used during the search to find the subsystem node
func (g *Graph) LoadSubsystemNodes(
	clusterName string,
	nodes *jgf.JsonGraph,
	subsystem string,
) error {

	// The graph needs to exist to add a subsystem to
	clusterG, ok := g.Clusters[clusterName]
	if !ok {
		return fmt.Errorf("cluster graph %s to register subsytem does not exist", clusterName)
	}
	return clusterG.LoadSubsystemNodes(nodes, subsystem)
}
