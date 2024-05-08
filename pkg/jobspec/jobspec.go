package jobspec

import (
	"fmt"
	"log"
	"strings"

	v1 "github.com/compspec/jobspec-go/pkg/nextgen/v1"
	rlog "github.com/converged-computing/rainbow/pkg/logger"
)

func JobspecFromCommand(command, jobName string, nodes, tasks int32) (*v1.Jobspec, error) {

	// Further validation of job happens with client below
	if command == "" {
		return &v1.Jobspec{}, fmt.Errorf("a command is required")
	}
	log.Printf("submit job: %s", command)

	// Prepare a JobSpec
	if jobName == "" {
		parts := strings.Split(command, " ")
		jobName = parts[0]
	}

	// Convert the simple command / nodes / etc into a JobSpec
	js, err := v1.NewSimpleJobspec(jobName, command, int32(nodes), int32(tasks))
	return js, err
}

// ShowRequires prints the requirements for a resource, if they exist
func ShowRequires(name string, resource *v1.Resource) {
	// Count will be defined for non slot, and replicas defined for slot
	count := resource.Count
	isSlot := ""
	if count == 0 {
		count = resource.Replicas
		isSlot = " (slot) "
	}
	rlog.Verbosef("     %s: %s %d\n", resource.Type, isSlot, count)
	if len(resource.Requires) > 0 {
		rlog.Verbosef("       requires\n")
		for _, entry := range resource.Requires {
			for key, value := range entry {
				rlog.Verbosef("         %s: %s\n", key, value)
			}
		}
	}

}
