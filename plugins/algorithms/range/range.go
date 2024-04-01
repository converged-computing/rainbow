package rangematch

// note range is a protected term

import (
	"fmt"
	"strings"

	semver "github.com/Masterminds/semver/v3"
	v1 "github.com/compspec/jobspec-go/pkg/jobspec/experimental"
	"github.com/converged-computing/rainbow/pkg/graph/algorithm"

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

var (
	description = "determine subsystem match based on membership in a range"
	matcherName = "range"
)

func (s RangeType) Name() string {
	return matcherName
}

func (s RangeType) Description() string {
	return description
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

// getSlotResource needs assumes a subsystem request as follows
/*
task:
  command:
  - spack
  slot: default
  count:
    per_slot: 1
  resources:
    spack:
      range:
      - field: version
        min: "0.5.1"
        max: "0.5.5"
*/
// We look for a field in the subsystem metadata attached to a node,
// in the example above "version" and then parse either > a min, < a max,
// or between the range.
func (m RangeType) GetSlotResourceNeeds(slot *v1.Task) *types.SlotResourceNeeds {
	sNeeds := map[string]map[string]bool{}
	for subsystem, needs := range slot.Resources {

		// Needs should be interface{} --> map[string][]map[string]string{}
		// Assume if we cannot parse, don't consider
		needs, ok := needs.(map[string]interface{})
		if !ok {
			continue
		}

		// Do we have a range algorithm?
		request, ok := needs["range"]
		if !ok {
			continue
		}

		// Now "request" goes from interface{} -> []map[string]string{}
		matches, ok := request.([]interface{})
		if !ok {
			continue
		}

		// Finally, we just parse the list - these should be key value pairs to match exactly
		for _, entry := range matches {
			entry, ok := entry.(map[string]interface{})
			if !ok {
				continue
			}

			// Go through each entry and parse into a request
			req := RangeRequest{}
			for key, value := range entry {
				value, ok := value.(string)

				// We only know how to parse these
				if key == "field" && ok {
					req.Field = value
				} else if key == "min" && ok {
					req.Min = value
				} else if key == "max" && ok {
					req.Max = value
				}
			}
			// If we get here and we have a field and at LEAST
			// one of min or max, we can add to to our needs
			// This is a bit janky - compressing with || separators
			if req.Field != "" && (req.Min != "" || req.Max != "") {
				_, ok := sNeeds[subsystem]
				if !ok {
					sNeeds[subsystem] = map[string]bool{}
				}
				// This sets the starting state that the range is not satisfied
				sNeeds[subsystem][req.Compress()] = false
			}
		}
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
func (m RangeType) CheckSubsystemEdge(
	slotNeeds *types.SlotResourceNeeds,
	edge *types.Edge,
	vtx *types.Vertex,
) {

	// Return early if we are satisfied
	if slotNeeds.Satisfied {
		return
	}

	// Determine if our slot needs can be met
	rlog.Debugf("Looking at edge %s->%s\n", edge.Relation, edge.Vertex.Type)

	// TODO Keep a record if all are satisfied so we stop searching
	// earlier if this is the case on subsequent calls
	for i, subsys := range slotNeeds.Subsystems {

		rlog.Debugf("      => Looking in subsystem %s\n", edge.Subsystem)

		// The subsystem has an edge defined here!
		if subsys.Name == edge.Subsystem {
			rlog.Debugf("      => Found matching subsystem %s for %s\n", subsys.Name, edge.Subsystem)

			// This would match the top level subsystem name
			for k := range subsys.Attributes {
				rlog.Debugf("      => Looking at edge %s '%s' for %s that needs %s\n", edge.Subsystem, edge.Vertex.Type, subsys.Name, k)

				// We care if the attribute is marked as a range
				if strings.HasPrefix(k, "range") {
					rlog.Debugf("      => Found %s and inspecting edge metadata %v\n", k, edge.Vertex.Metadata.Elements)
					req := NewRangeRequest(k)

					// Get the field requested by the jobspec
					toMatch, err := edge.Vertex.Metadata.GetStringElement(req.Field)
					if err != nil {
						continue
					}

					rlog.Debugf("      => Found field requested for range match %s\n", toMatch)
					satisfied, err := req.Satisfies(toMatch)
					if err != nil {
						continue
					}
					if satisfied {
						rlog.Debugf("      => Resource '%s' has edge '%s' satisfies subsystem %s %s\n", vtx.Type, edge.Vertex.Type, subsys.Name, k)
						subsys.Attributes[k] = true
					}
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
func (s RangeType) Init(options map[string]string) error {
	// If an algorithm has options, they can be set here
	return nil
}

// Add the selection algorithm to be known to rainbow
func init() {
	algo := RangeType{}
	algorithm.Register(algo)
}
