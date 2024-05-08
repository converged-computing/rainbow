package equals

import (
	"fmt"
	"strings"

	"github.com/converged-computing/rainbow/pkg/graph/algorithm"
	rlog "github.com/converged-computing/rainbow/pkg/logger"
	"github.com/converged-computing/rainbow/pkg/types"
)

type EqualsType struct{}

var (
	description = "simplest match that only allows exact equality"
	matcherName = "equals"
)

// A match request can be for a range or exact match
type MatchEqualRequest struct {
	Field string
	Value string
}

func (s EqualsType) Name() string {
	return matcherName
}

func (s EqualsType) Description() string {
	return description
}

// Compress the match request into a parseable field
func (req *MatchEqualRequest) Compress() string {
	value := fmt.Sprintf("match||field=%s", req.Field)
	value = fmt.Sprintf("%s||value=%s", value, req.Value)
	return value
}

func NewMatchEqualRequest(value string) *MatchEqualRequest {
	req := MatchEqualRequest{}
	pieces := strings.Split(value, "||")
	for _, piece := range pieces {
		if strings.HasPrefix(piece, "field=") {
			req.Field = strings.ReplaceAll(piece, "field=", "")
		} else if strings.HasPrefix(piece, "value=") {
			req.Value = strings.ReplaceAll(piece, "value=", "")
		}
	}
	return &req
}

// MatchEdge is an exposed function (for other matchers to use)
// to allow for matching a subsystem edge
func (m EqualsType) MatchEdge(k string, edge *types.Edge, subsys *types.SubsystemNeeds) {
	req := NewMatchEqualRequest(k)
	// Get the field requested by the jobspec
	toMatch, err := edge.Vertex.Metadata.GetStringElement(req.Field)
	if err != nil {
		return
	}

	rlog.Debugf("      => Found field requested for range match %s\n", toMatch)
	// These are the conditions of being satisifed, the value we got from the vertex
	// matches the value provided in the slot request
	if toMatch == req.Value {
		rlog.Debugf("      => Edge '%s' satisfies subsystem %s %s\n", edge.Vertex.Type, subsys.Name, k)
		subsys.Attributes[k] = true
	}
}

// UpdateResourceNeeds take a match interface and updates resource needs
// This is provided to expose the match interface to other matches
func UpdateResourceNeeds(
	needs map[string]string,
	subsystem string,
	sNeeds map[string]map[string]bool,
) map[string]map[string]bool {

	// Finally, we just parse the list - these should be key value pairs to match exactly
	req := MatchEqualRequest{}

	// Cut out early if not a match
	match, ok := needs["match"]
	if !ok {
		return sNeeds
	}

	req.Value = match
	field, ok := needs["field"]
	if !ok {
		return sNeeds
	}
	req.Field = field
	if req.Field != "" && req.Value != "" {
		_, ok := sNeeds[subsystem]
		if !ok {
			sNeeds[subsystem] = map[string]bool{}
		}
		// This sets the starting state that the range is not satisfied
		sNeeds[subsystem][req.Compress()] = false
	}
	return sNeeds
}

// checkSubsystemEdge evaluates a node edge in the dominant subsystem for a
// subsystem attribute. E.g., if the io subsystem provides
// Vertex (from dominant subsysetem) is only passed in for informational purposes
func (m EqualsType) CheckSubsystemEdge(slotNeeds *types.SlotResourceNeeds, edge *types.Edge, vtx *types.Vertex) {

	// Return early if we are satisfied
	if slotNeeds.Satisfied {
		return
	}
	// Determine if our slot needs can be met
	// Nested for loops are not great - this will be improved with a more robust graph
	// that isn't artisinal avocado toast developed by me :)

	rlog.Debugf("Looking at edge %s->%s\n", edge.Relation, edge.Vertex.Type)

	// TODO Keep a record if all are satisfied so we stop searching
	// earlier if this is the case on subsequent calls
	for i, subsys := range slotNeeds.Subsystems {

		rlog.Debugf("      => Looking in subsystem %s\n", edge.Subsystem)

		// The subsystem has an edge defined here!
		if subsys.Name == edge.Subsystem {
			rlog.Debugf("      => Found matching subsystem %s for %s\n", subsys.Name, edge.Subsystem)

			// Yuck, this needs to be a query! Oh well.
			for k := range subsys.Attributes {
				rlog.Debugf("      => Looking at edge %s '%s' for %s that needs %s\n", edge.Subsystem, edge.Vertex.Type, subsys.Name, k)

				// We care if the attribute is marked as a range
				if strings.HasPrefix(k, "match") {
					m.MatchEdge(k, edge, &subsys)
				}
			}
		}
		slotNeeds.Subsystems[i] = subsys
	}

	// Try to avoid future checking if subsystem needs are addressed
	allSatisfied := true
	for _, subsys := range slotNeeds.Subsystems {
		for _, v := range subsys.Attributes {
			if !v {
				allSatisfied = false
				break
			}
		}
	}
	// This is going to provide a quick check to determine if the subsystem
	// is satisfied without needing to parse again
	slotNeeds.Satisfied = allSatisfied
}

// Init provides extra initialization functionality, if needed
// The in memory database can take a backup file if desired
func (s EqualsType) Init(options map[string]string) error {
	// If an algorithm has options, they can be set here
	return nil
}

// Add the selection algorithm to be known to rainbow
func init() {
	algo := EqualsType{}
	algorithm.Register(algo)
}
