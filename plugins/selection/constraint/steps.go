package constraint

import (
	js "github.com/compspec/jobspec-go/pkg/jobspec/experimental"

	"github.com/Knetic/govaluate"
	"github.com/converged-computing/rainbow/pkg/types"
)

// filterStep applies a filter to the list of clusters
func filterStep(
	clusters *[]string,
	logic string,
	states map[string]types.ClusterState,
	jobspec *js.Jobspec,
) ([]string, error) {

	filtered := []string{}

	// The expression is formed from the logic provided
	expression, err := govaluate.NewEvaluableExpression(logic)
	if err != nil {
		return filtered, err
	}
	for _, cluster := range *clusters {

		// Parameters are populated from each cluster state and the jobpsec
		// This is definitely not going to be efficient
		totalParams := len(jobspec.Attributes.Parameter) + len(states[cluster])
		parameters := make(map[string]interface{}, totalParams)
		if jobspec.Attributes.Parameter != nil {

			// Add all parameters we have to be available
			for key, value := range jobspec.Attributes.Parameter {
				parameters[key] = value
			}

			// Add metadata for the cluster
			for key, value := range states[cluster] {
				parameters[key] = value
			}
		}

		// Try to evaluate the expression
		passes, err := expression.Evaluate(parameters)
		if passes == "false" || err != nil {
			continue
		}
		filtered = append(filtered, cluster)
	}
	return filtered, nil
}
