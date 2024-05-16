package match

import (
	"fmt"
	"strings"

	semver "github.com/Masterminds/semver/v3"

	rlog "github.com/converged-computing/rainbow/pkg/logger"
	"github.com/converged-computing/rainbow/pkg/types"
)

type RangeType struct{}
type RangeRequest struct {
	Min   string
	Max   string
	Field string
}

// Compress into a string to hand off to the graph for later matching
func (req *RangeRequest) Compress() string {

	value := fmt.Sprintf("range||field=%s", req.Field)
	if req.Min != "" {
		value = fmt.Sprintf("%s||min=%s", value, req.Min)
	}
	if req.Max != "" {
		value = fmt.Sprintf("%s||max=%s", value, req.Max)
	}
	return value
}

func NewRangeRequest(value string) *RangeRequest {
	req := RangeRequest{}
	pieces := strings.Split(value, "||")
	for _, piece := range pieces {
		if strings.HasPrefix(piece, "min=") {
			req.Min = strings.ReplaceAll(piece, "min=", "")
		} else if strings.HasPrefix(piece, "max=") {
			req.Max = strings.ReplaceAll(piece, "max=", "")
		} else if strings.HasPrefix(piece, "field=") {
			req.Field = strings.ReplaceAll(piece, "field=", "")
		}
	}
	return &req
}

// GetResourceNeeds of a range request
func (r *RangeRequest) GetResourceNeeds(request map[string]string) map[string]bool {
	needs := map[string]bool{}
	for key, value := range request {

		// We only know how to parse these
		if key == "field" {
			r.Field = value
		} else if key == "min" {
			r.Min = value
		} else if key == "max" {
			r.Max = value
		}
	}
	// If we get here and we have a field and at LEAST
	// one of min or max, we can add to to our needs
	// This is a bit janky - compressing with || separators
	if r.Field != "" && (r.Min != "" || r.Max != "") {
		needs[r.Compress()] = false
	}
	return needs
}

// Determine if a range request satisfies the node field
// If the user specifies a wonky range, this will still work,
// but not as they expect :)
func (req *RangeRequest) Satisfies(value string) (bool, error) {

	// We already have the value for the field from the graph, now just use semver to match
	matchVersion, err := semver.NewVersion(value)
	if err != nil {
		rlog.Debugf("      => Error parsing semver for match value %s\n", err)
		return false, err
	}
	if req.Min != "" {
		// Is the version provided greater than the min requested?
		c, err := semver.NewConstraint(fmt.Sprintf(">= %s", req.Min))
		if err != nil {
			rlog.Debugf("      => Error parsing min constraint %s\n", err)
			return false, err
		}
		// Check if the version meets the constraints. The a variable will be true.
		satisfied := c.Check(matchVersion)
		if !satisfied {
			rlog.Debugf("      => Not satisfied\n")
			return false, err

		}
	}
	if req.Max != "" {
		// Is the version provided less than the max requested?
		c, err := semver.NewConstraint(fmt.Sprintf("<= %s", req.Max))
		if err != nil {
			rlog.Debugf("      => Error parsing max constraint %s\n", err)
			return false, err
		}
		// Check if the version meets the constraints. The a variable will be true.
		satisfied := c.Check(matchVersion)
		if !satisfied {
			rlog.Debug("      => Not satisfied")
			return false, err
		}
	}
	return true, nil
}

// MatchRangeEdge matches to a range
func MatchRangeEdge(matchExpression string, edge *types.Edge) bool {
	rlog.Debugf("      => Found %s and inspecting edge metadata %v\n", matchExpression, edge.Vertex.Metadata.Elements)
	req := NewRangeRequest(matchExpression)

	// Get the field requested by the jobspec
	toMatch, err := edge.Vertex.Metadata.GetStringElement(req.Field)
	if err != nil {
		return false
	}

	rlog.Debugf("      => Found field requested for range match %s\n", toMatch)
	satisfied, err := req.Satisfies(toMatch)
	if err != nil {
		return false
	}
	return satisfied
}

// MatchEqualityCypher writes the lines of cypher for a match
func MatchRangeCypher(subsystem, matchExpression string) string {
	req := NewRangeRequest(matchExpression)

	// req.Name => the subsystem
	query := fmt.Sprintf("\n-[contains]-(%s:Node {subsystem: '%s'})", subsystem, subsystem)

	// Need to assemble min/max, or both
	queryPiece := "\nWHERE"
	if req.Min != "" {
		queryPiece += fmt.Sprintf("%s.%s >= %d", subsystem, req.Field, req.Min)
	}
	if req.Max != "" {
		queryPiece += fmt.Sprintf("AND %s.%s <= %d", subsystem, req.Field, req.Max)
	}
	query += queryPiece
	return query
}
