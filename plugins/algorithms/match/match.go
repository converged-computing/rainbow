package match

import (
	"strings"

	"github.com/converged-computing/rainbow/pkg/graph/algorithm"
	rlog "github.com/converged-computing/rainbow/pkg/logger"
	"github.com/converged-computing/rainbow/pkg/types"
)

type MatchType struct{}

var (
	description = "match single values or ranges for subsystem job assignment"
	matcherName = "match"
)

func (s MatchType) Name() string {
	return matcherName
}

func (s MatchType) Description() string {
	return description
}

// checkSubsystemNeeds is a shared function to loop over needs for an edge to check
// the needs should already be scoepd to a subsystem
func CheckSubsystemNeeds(needs map[string]bool, edge *types.Edge) map[string]bool {

	// Yuck, this would be better as a query! Oh well.
	for k := range needs {
		rlog.Debugf("      => Looking at edge %s '%s' for %s that needs %s\n", edge.Subsystem, edge.Vertex.Type, edge.Subsystem, k)

		// We care if the attribute is marked as a range or exact match
		// Passing slotNeeds to the function (pointer) updates in place
		if strings.HasPrefix(k, "match") {
			if MatchEqualityEdge(k, edge) {
				rlog.Debugf("      => Edge '%s' satisfies subsystem %s %s\n", edge.Vertex.Type, edge.Subsystem, k)
				needs[k] = true
			}

		} else if strings.HasPrefix(k, "range") {
			if MatchRangeEdge(k, edge) {
				rlog.Debugf("      => Edge '%s' satisfies subsystem %s %s\n", edge.Vertex.Type, edge.Subsystem, k)
				needs[k] = true
			}
		}
	}
	return needs
}

// checkSubsystemEdge evaluates a node edge in the dominant subsystem for a
// subsystem attribute. E.g., if the io subsystem provides
// Vertex (from dominant subsysetem) is only passed in for informational purposes
func (m MatchType) CheckSubsystemEdge(
	slotNeeds *types.MatchAlgorithmNeeds,
	edge *types.Edge,
) {

	// Determine if our slot needs can be met
	// Nested for loops are not great - this will be improved with a more robust graph
	// that isn't artisinal avocado toast developed by me :)

	rlog.Debugf("Looking at edge %s->%s\n", edge.Relation, edge.Vertex.Type)

	updatedNeeds := types.MatchAlgorithmNeeds{}

	// TODO Keep a record if all are satisfied so we stop searching
	// earlier if this is the case on subsequent calls
	for subsystem, needs := range *slotNeeds {

		// The subsystem has an edge defined here!
		if subsystem == edge.Subsystem {
			rlog.Debugf("      => Found subsystem %s edge to search\n", edge.Subsystem)
			needs = CheckSubsystemNeeds(needs, edge)
		}
		updatedNeeds[subsystem] = needs
	}
	// Update the slot needs that get passed back
	slotNeeds = &updatedNeeds
}

// GetResourceNeeds of a match request
func (r *MatchEqualRequest) GetResourceNeeds(request map[string]string) map[string]bool {
	needs := map[string]bool{}

	// Cut out early if not a match
	match, ok := request["match"]
	if !ok {
		return needs
	}
	r.Value = match
	field, ok := request["field"]
	if !ok {
		return needs
	}
	r.Field = field
	if r.Field != "" && r.Value != "" {
		// This sets the starting state that the range is not satisfied
		needs[r.Compress()] = false
	}
	return needs
}

// UpdateResourceNeeds allows exposing parsing of the match interface
// to other matchers
func GetResourceNeeds(request map[string]string) map[string]bool {

	// Go through each entry and parse into a request
	// In practice, there should only be one of these set at a time
	req := RangeRequest{}
	needs := req.GetResourceNeeds(request)
	eq := MatchEqualRequest{}
	for key, value := range eq.GetResourceNeeds(request) {
		needs[key] = value
	}
	return needs
}

// Init provides extra initialization functionality, if needed
// The in memory database can take a backup file if desired
func (s MatchType) Init(options map[string]string) error {
	// If an algorithm has options, they can be set here
	return nil
}

// Generate cypher for the match algorithm for a specific slot
func (m MatchType) GenerateCypher(matchNeeds *types.MatchAlgorithmNeeds) string {

	// This will be added as a piece in a query we are building
	query := ""
	for subsystemName, needs := range *matchNeeds {

		// k is the string to parse, we can assume since we do one query
		// that the boolean is always false
		for matchExpression := range needs {
			if strings.HasPrefix(matchExpression, "match") {
				query += MatchEqualityCypher(subsystemName, matchExpression)

			} else if strings.HasPrefix(matchExpression, "range") {
				query += MatchRangeCypher(subsystemName, matchExpression)
			}
		}
	}

	return query
}

// Add the selection algorithm to be known to rainbow
func init() {
	algo := MatchType{}
	algorithm.Register(algo)
}
