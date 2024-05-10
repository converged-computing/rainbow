package constraint

import (
	"math"
	"math/rand"
	"sort"
	"strings"

	js "github.com/compspec/jobspec-go/pkg/nextgen/v1"

	"github.com/Knetic/govaluate"
	rlog "github.com/converged-computing/rainbow/pkg/logger"
	"github.com/converged-computing/rainbow/pkg/types"
)

// Common functions likely desired for govaluate
var (
	functions = map[string]govaluate.ExpressionFunction{
		"min": func(args ...interface{}) (interface{}, error) {
			valA := args[0].(float64)
			valB := args[1].(float64)
			return math.Min(valA, valB), nil
		},
		"max": func(args ...interface{}) (interface{}, error) {
			valA := args[0].(float64)
			valB := args[1].(float64)
			return math.Max(valA, valB), nil
		},
	}
)

// ClusterSort is for cluster sorting
type ClusterSort struct {
	Name  string
	State *types.ClusterState
	// For now we only support int32 comparisons
	Value int32
}

// filterStep applies a filter to the list of clusters
func filterStep(
	clusters *[]string,
	logic string,
	states map[string]types.ClusterState,
	jobspec *js.Jobspec,
) ([]string, error) {

	filtered := []string{}

	// The expression is formed from the logic provided
	rlog.Debugf("      checking if %s\n", logic)
	expression, err := govaluate.NewEvaluableExpressionWithFunctions(logic, functions)
	if err != nil {
		return filtered, rlog.ErrorPrintf("    logic expression %s is not valid: %s", err)
	}
	for _, cluster := range *clusters {

		// get parameters from the jobspec and cluster state
		state := states[cluster]
		parameters := prepareParameters(jobspec, &state)

		// Try to evaluate the expression
		passes, err := expression.Evaluate(parameters)
		if passes == "false" {
			continue
		}
		if err != nil {
			rlog.Warningf("    issue with filter evaluation: %s\n", err)
			continue
		}
		filtered = append(filtered, cluster)
	}
	return filtered, nil
}

func calcStep(
	logic string,
	states map[string]types.ClusterState,
	jobspec *js.Jobspec,
) (map[string]types.ClusterState, error) {

	// The logic expression must have two parts
	parts := strings.Split(logic, "=")
	if len(parts) != 2 {
		return states, rlog.ErrorPrintf("    logic expression %s is not valid, should be in format var=expr", logic)
	}
	varname, logic := parts[0], parts[1]

	// The expression is formed from the logic provided
	expression, err := govaluate.NewEvaluableExpressionWithFunctions(logic, functions)
	if err != nil {
		return states, rlog.ErrorPrintf("    logic expression %s is not valid: %s", err)
	}

	for cluster, state := range states {

		// get parameters from the jobspec and cluster state
		parameters := prepareParameters(jobspec, &state)

		// Try to evaluate the expression
		result, err := expression.Evaluate(parameters)
		if err != nil {
			rlog.Warningf("    issue with calc evaluation: %s\n", err)
			continue
		}
		// Update cluster states to include calculate value
		rlog.Debugf("      adding calculated value %s:%v to states for cluster %s\n", varname, result, cluster)
		states[cluster][varname] = result
	}
	return states, nil
}

// finalSelect applies a method (e.g., random) to make a final choice
func finalSelect(
	clusters []string,
	method string,
) (string, error) {

	if method == "random" {
		idx := rand.Intn(len(clusters))
		return clusters[idx], nil
	}
	if method == "last" {
		return clusters[len(clusters)-1], nil
	}
	if method == "first" {
		return clusters[0], nil
	}
	return "", rlog.ErrorPrintf("    %s is not a valid select method", method)
}

// sortDescending/ascending sort the clusters based on a parameter
func sortDescending(
	variable string,
	states map[string]types.ClusterState,
	jobspec *js.Jobspec,
) ([]string, error) {

	lookup := *generateClusterLookup(variable, states, jobspec)

	// Ensure we sort by priority value, just once
	sort.Slice(lookup, func(i, j int) bool {
		return lookup[i].Value < lookup[j].Value
	})

	// Assemble back into list
	final := []string{}
	for _, entry := range lookup {
		final = append(final, entry.Name)
	}
	return final, nil
}

// sortDescending/ascending sort the clusters based on a parameter
func sortAscending(
	variable string,
	states map[string]types.ClusterState,
	jobspec *js.Jobspec,
) ([]string, error) {

	lookup := *generateClusterLookup(variable, states, jobspec)

	// Ensure we sort by priority value, just once
	sort.Slice(lookup, func(i, j int) bool {
		return lookup[i].Value > lookup[j].Value
	})

	// Assemble back into list
	final := []string{}
	for _, entry := range lookup {
		final = append(final, entry.Name)
	}
	return final, nil
}

// prepareParameters is a shared function that steps can use for equation parameters
func prepareParameters(jobspec *js.Jobspec, state *types.ClusterState) map[string]interface{} {

	totalParams := len(*state)
	parameters := make(map[string]interface{}, totalParams)

	// Parameters are populated from each cluster state and the jobpsec
	// This is definitely not going to be efficient
	parameter, ok := jobspec.Attributes["parameter"]
	if ok {

		// This is a bit dangerous - could panic if wrong type provided
		params := parameter.(js.Attributes)

		totalParams = len(params) + len(*state)
		parameters = make(map[string]interface{}, totalParams)

		// Add all parameters we have to be available
		for key, value := range params {
			rlog.Debugf("      adding parameter %s=%v\n", key, value)
			parameters[key] = value
		}
	}
	// Add metadata for the cluster
	for key, value := range *state {
		rlog.Debugf("      adding parameter %s=%v\n", key, value)
		parameters[key] = value
	}
	return parameters
}

// generateClusterLookup is a helper for either sort function
func generateClusterLookup(
	variable string,
	states map[string]types.ClusterState,
	jobspec *js.Jobspec,
) *[]ClusterSort {

	// Prepare lookup of parameters, we need this for sorting
	lookup := make([]ClusterSort, len(states))
	for cluster, state := range states {
		parameters := prepareParameters(jobspec, &state)

		// If we don't have the variable, don't include it
		value, ok := parameters[variable]
		if !ok {
			continue
		}
		switch value.(type) {
		case int32:
			break
		default:
			continue
		}
		lookup = append(lookup, ClusterSort{
			Name:  cluster,
			State: &state,
			Value: value.(int32),
		})
	}
	return &lookup
}
