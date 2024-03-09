package jobspec

import (
	"fmt"
	"log"
	"strings"

	js "github.com/compspec/jobspec-go/pkg/jobspec/experimental"
)

func JobspecFromCommand(command, jobName string, nodes, tasks int32) (*js.Jobspec, error) {

	// Further validation of job happens with client below
	if command == "" {
		return &js.Jobspec{}, fmt.Errorf("a command is required")
	}
	log.Printf("submit job: %s", command)

	// Prepare a JobSpec
	if jobName == "" {
		parts := strings.Split(command, " ")
		jobName = parts[0]
	}

	// Convert the simple command / nodes / etc into a JobSpec
	js, err := js.NewSimpleJobspec(jobName, command, int32(nodes), int32(tasks))
	return js, err
}
