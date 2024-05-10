package match

import (
	"fmt"
	"strings"

	rlog "github.com/converged-computing/rainbow/pkg/logger"
	"github.com/converged-computing/rainbow/pkg/types"
)

// Equality check (an exact match)

type EqualsType struct{}
type MatchEqualRequest struct {
	Field string
	Value string
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

// MatchEqualityEdge looks for an exact match
func MatchEqualityEdge(k string, edge *types.Edge) bool {
	req := NewMatchEqualRequest(k)

	// Get the field requested by the jobspec
	toMatch, err := edge.Vertex.Metadata.GetStringElement(req.Field)
	if err != nil {
		return false
	}

	rlog.Debugf("      => Found field requested for range match %s\n", toMatch)
	// These are the conditions of being satisifed, the value we got from the vertex
	// matches the value provided in the slot request
	return toMatch == req.Value
}
