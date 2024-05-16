package graph

import (
	"encoding/json"
	"fmt"
	"os"

	v1 "github.com/compspec/jobspec-go/pkg/nextgen/v1"

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

type SlotCount struct {

	// The number of slots required
	Count int32
	Name  string

	// The number of resource submembers needed per slot
	Members int32

	// The parent of the slot
	Parent string
}

// ExtractResourceSlots flattens a jobspec into a lookup of slots
func ExtractResourceSlots(jobspec *v1.Jobspec) []SlotCount {

	totals := []SlotCount{}

	// Go sets loops to an initial value at start,
	// so we need a function to recurse into nested resources
	var checkResource func(resource *v1.Resource)
	checkResource = func(resource *v1.Resource) {
		// Assume a slot is a count for 1 resource type
		// If we find the slot, we go just below it
		// We just need the total for the slot level
		if resource.Replicas != 0 {

			// This is the recursive bit
			if resource.With != nil {
				for _, with := range resource.With {
					newSlot := SlotCount{
						Count:   resource.Replicas,
						Name:    with.Type,
						Members: with.Count,
						Parent:  resource.Type}
					totals = append(totals, newSlot)
				}
			}
		} else {
			if resource.With != nil {
				for _, with := range resource.With {
					checkResource(&with)
				}
			}
		}
	}
	// Make a call on each of the top level resources
	for _, resource := range jobspec.Resources {
		checkResource(&resource)
	}
	return totals
}

// A Slot for the slotlist
type Slot struct {
	Name  string
	Count int32
}

// ExtractResourceSlots flattens a jobspec into slots, but ordered by appearance
func ExtractResourceSlotList(jobspec *v1.Jobspec) []Slot {

	totals := map[string]Slot{}

	// Go sets loops to an initial value at start,
	// so we need a function to recurse into nested resources
	var checkResource func(resource *v1.Resource)
	checkResource = func(resource *v1.Resource) {
		slot, ok := totals[resource.Type]
		if !ok {
			slot = Slot{Count: 0}
		}
		slot.Count += resource.Count
		totals[resource.Type] = slot

		// This is the recursive bit
		if resource.With != nil {
			for _, with := range resource.With {
				checkResource(&with)
			}
		}
	}
	// Make a call on each of the top level resources
	for _, resource := range jobspec.Resources {
		checkResource(&resource)
	}

	// Turn into a list
	slots := []Slot{}
	for _, slot := range totals {
		slots = append(slots, slot)
	}
	return slots
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
