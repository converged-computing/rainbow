package constraint

import (
	"fmt"
	"sort"

	js "github.com/compspec/jobspec-go/pkg/jobspec/experimental"
	"github.com/converged-computing/rainbow/pkg/graph/selection"
	rlog "github.com/converged-computing/rainbow/pkg/logger"
	"github.com/converged-computing/rainbow/pkg/types"
	"github.com/converged-computing/rainbow/pkg/utils"
	"gopkg.in/yaml.v3"
)

var (
	opts = []ConstraintPriority{}
)

/* Constraint selection of a cluster.
Here the algorithm takes the following approach:
Provide a list of priority filters. Each can include a series of steps to:
- filter
- calculate (calc)
- sort_descending / sort_ascending
- select (random, first, last)

We go through each priority once, and if there are no results left,
try the next in the list until we run out
*/

type ConstraintSelection struct{}

var (
	description  = "selection based on prioritized constraints"
	selectorName = "constraint"
)

func (s ConstraintSelection) Name() string {
	return selectorName
}

func (s ConstraintSelection) Description() string {
	return description
}

// Select randomly chooses a cluster from the set
// This should not receive an empty list, but we check anyway
func (s ConstraintSelection) Select(
	contenders []string,
	states map[string]types.ClusterState,
	jobspec string,
) (string, error) {
	if len(contenders) == 0 {
		return "", nil
	}
	rlog.Debugf("  Pre-state filter clusters: %s\n", contenders)

	// This algorithm requires that we have state data for matches
	matches := []string{}
	for _, cluster := range contenders {
		_, ok := states[cluster]
		if ok {
			matches = append(matches, cluster)
		}
	}
	rlog.Debugf("  Post-state filter clusters: %s\n", matches)

	// Again, cut out early with no match
	if len(matches) == 0 {
		return "", nil
	}

	// Load the jobspec from yaml string
	jspec := js.Jobspec{}
	err := yaml.Unmarshal([]byte(jobspec), &jspec)
	if err != nil {
		return "", err
	}

	// Loop through priorities until we have a match (or finish and no match)
	// Note this is implemented to work - I haven't thought about optimizing it
	for _, priority := range opts {

		// Copy contenders
		clusters := utils.Copy(matches)

		// TODO we might also want to copy states, as calc
		// for a priority level can change it. I'm assuming now that
		// if the change happens, it will overwrite an existing or
		// simply create a new variable name
		rlog.Debugf("    priority %d: %d initial clusters\n", priority.Priority, len(clusters))

		// indicator to tell us to break from two loops
		nextPriority := false
		for _, steps := range priority.Steps {
			if nextPriority {
				break
			}

			// Go through steps until we get matches (or need to bail out)
			for stepName, logic := range steps {

				// A "filter" step will use provided logic, jobspec, and states to filter clusters
				if stepName == "filter" {
					clusters, err = filterStep(&clusters, logic, states, &jspec)
					if err != nil {
						return "", err
					}
					rlog.Debugf("    filter: %d clusters remaining\n", len(clusters))

					// If no matches after the filter, we need to continue
					if len(clusters) == 0 {
						nextPriority = true
						break
					}
				} else if stepName == "calc" {
					rlog.Debugf("    calc: %s\n", logic)
					states, err = calcStep(logic, states, &jspec)
					if err != nil {
						return "", err
					}
				} else if stepName == "sort_descending" {
					clusters, err := sortDescending(logic, states, &jspec)
					if err != nil {
						return "", err
					}
					rlog.Debugf("    sort_descending: %d clusters\n", len(clusters))
					if len(clusters) == 0 {
						nextPriority = true
						break
					}
				} else if stepName == "sort_ascending" {
					clusters, err := sortAscending(logic, states, &jspec)
					if err != nil {
						return "", err
					}
					rlog.Debugf("    sort_ascending: %d clusters\n", len(clusters))
					if len(clusters) == 0 {
						nextPriority = true
						break
					}
				} else if stepName == "select" {
					cluster, err := finalSelect(clusters, logic)
					if err != nil {
						return "", err
					}
					// We found a match!
					if cluster != "" {
						rlog.Debugf("    select: clusters %s is selected\n", cluster)
						return cluster, nil
					}
				}
			}
		}
	}
	// If we get here, there were no matches
	return "", nil
}

// Init provides extra initialization functionality, if needed
// The in memory database can take a backup file if desired
func (s ConstraintSelection) Init(options map[string]string) error {
	// This algorithm requires priorities to be set
	priorities, ok := options["priorities"]
	if !ok {
		return fmt.Errorf("the constraint selection algorithm requires priorities to be defined in options")
	}
	// Parse into global options for later use
	err := yaml.Unmarshal([]byte(priorities), &opts)
	if err != nil {
		return err
	}
	// Ensure we sort by priority value, just once
	sort.Slice(opts, func(i, j int) bool {
		return opts[i].Priority < opts[j].Priority
	})
	return nil
}

// Add the selection algorithm to be known to rainbow
func init() {
	algo := ConstraintSelection{}
	selection.Register(algo)
}
