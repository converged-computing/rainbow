package graph

import (
	"encoding/json"
	"fmt"
	"os"

	jgf "github.com/converged-computing/jsongraph-go/jsongraph/v2/graph"

	"github.com/converged-computing/jsongraph-go/jsongraph/v2/graph"
)

// ReadNodeJsonGraph reads in the node JGF
// We read it in just to validate, but serialize as string
func ReadNodeJsonGraph(jsonFile string) (graph.JsonGraph, string, error) {

	g := graph.JsonGraph{}

	file, err := os.ReadFile(jsonFile)
	if err != nil {
		return g, "", fmt.Errorf("error reading %s:%s", jsonFile, err)
	}

	err = json.Unmarshal([]byte(file), &g)
	if err != nil {
		return g, "", fmt.Errorf("error unmarshalling %s:%s", jsonFile, err)
	}
	return g, string(file), nil
}

func ReadNodeJsonGraphString(nodes string) (graph.JsonGraph, error) {
	g := graph.JsonGraph{}
	err := json.Unmarshal([]byte(nodes), &g)
	if err != nil {
		return g, fmt.Errorf("error unmarshalling json graph: %s", err)
	}
	return g, nil
}

// GetNamespacedName is a shared function to get a namespaced name for a node/edge
func GetNamespacedName(clusterName, name string) string {
	return fmt.Sprintf("%s-%s", clusterName, name)
}

// validateNodes ensures that we have at least one node and edge
func ValidateNodes(nodes *jgf.JsonGraph) (int, int, error) {
	var err error
	nNodes := len(nodes.Graph.Nodes)
	nEdges := len(nodes.Graph.Edges)
	if nEdges == 0 || nNodes == 0 {
		err = fmt.Errorf("subsystem cluster must have at least one edge and node")
	}
	return nNodes, nEdges, err
}
