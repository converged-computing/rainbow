package equals

import (
	"fmt"
	"strings"

	v1 "github.com/compspec/jobspec-go/pkg/jobspec/experimental"
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
	rlog.Debugf("      => Found %s and inspecting edge metadata %v\n", k, edge.Vertex.Metadata.Elements)
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
func (m EqualsType) UpdateResourceNeeds(
	match interface{},
	subsystem string,
	sNeeds map[string]map[string]bool,
) map[string]map[string]bool {

	// Now "match" goes from interface{} -> []map[string]string{}
	matches, ok := match.([]interface{})
	if !ok {
		return sNeeds
	}

	// Finally, we just parse the list - these should be key value pairs to match exactly
	for _, entry := range matches {
		entry, ok := entry.(map[string]interface{})
		if !ok {
			continue
		}

		req := MatchEqualRequest{}
		for key, value := range entry {
			value, ok := value.(string)

			// We only know how to parse these
			if key == "field" && ok {
				req.Field = value
			} else if key == "value" && ok {
				req.Value = value
			}
		}

		// If we get here and we have a field and at LEAST
		// one of min or max, we can add to to our needs
		// This is a bit janky - compressing with || separators
		if req.Field != "" && (req.Value != "") {
			_, ok := sNeeds[subsystem]
			if !ok {
				sNeeds[subsystem] = map[string]bool{}
			}
			// This sets the starting state that the range is not satisfied
			sNeeds[subsystem][req.Compress()] = false
		}
	}
	return sNeeds
}

// getSlotResource needs assumes a subsystem request as follows:
/* task:
command:
- ior
  slot: default
  count:
  per_slot: 1
resources:
  io:
  match:
  - type: shm
*/
// it is an explicit match, so we expect the slot to have that exact resource
// available. This can eventually take a count, but right now is a boolean match
// and this is done intentionally to satisfy the simplest scheduler experiment
// prototype where we are more interested in features
func (m EqualsType) GetSlotResourceNeeds(slot *v1.Task) *types.SlotResourceNeeds {
	sNeeds := map[string]map[string]bool{}
	for subsystem, needs := range slot.Resources {

		// Needs should be interface{} --> map[string][]map[string]string{}
		// Assume if we cannot parse, don't consider
		needs, ok := needs.(map[string]interface{})
		if !ok {
			continue
		}

		// We currently support "match" which is an exact match of a term to resource
		match, ok := needs["match"]
		if !ok {
			continue
		}
		sNeeds = m.UpdateResourceNeeds(match, subsystem, sNeeds)
	}
	// Parse into the slot resource needs
	needs := []types.SubsystemNeeds{}
	for subsystem, sneeds := range sNeeds {
		subsystemNeeds := types.SubsystemNeeds{Name: subsystem, Attributes: sneeds}
		needs = append(needs, subsystemNeeds)
	}

	// If we don't have any needs, the slot is satisfied for that
	slotNeeds := &types.SlotResourceNeeds{Subsystems: needs}
	if len(needs) == 0 {
		slotNeeds.Satisfied = true
	}
	rlog.Debugf("      => Assessing needs for slot: %v\n", slotNeeds)
	return slotNeeds
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
