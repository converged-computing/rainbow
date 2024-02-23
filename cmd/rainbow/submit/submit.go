package match

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/converged-computing/rainbow/pkg/client"
	"github.com/converged-computing/rainbow/pkg/types"
)

// Run will check a manifest list of artifacts against a host machine
// For now, the host machine parameters will be provided as flags
func Run(
	host, jobName, command string,
	nodes, tasks int,
	token, clusterName, cfgFile string,
) error {

	c, err := client.NewClient(host)
	if err != nil {
		return nil
	}

	// Further validation of job happens with client below
	if command == "" {
		return fmt.Errorf("a command is required")
	}
	log.Printf("submit job: %s", command)

	// Prepare a JobSpec
	if jobName == "" {
		parts := strings.Split(command, " ")
		jobName = parts[0]
	}
	jobspec := types.JobSpec{
		Name:    jobName,
		Nodes:   int32(nodes),
		Tasks:   int32(tasks),
		Command: command,
	}

	// Read in the config, if provided, TODO we need a set of tokens here?
	//cfg, err := config.NewRainbowClientConfig(cfgFile, clusterName, "")
	//if err != nil {
	//	return err
	//}

	// Last argument is secret, empty for now
	response, err := c.SubmitJob(context.Background(), jobspec, clusterName, token)
	if err != nil {
		return err
	}

	// If we get here, success! Dump all the stuff.
	//log.Printf("status: %s", response.Status)
	//log.Printf(" token: %s", response.Token)
	log.Println(response)
	return nil
}
