package graph

import (
	"encoding/json"
	"fmt"
	"os"

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
