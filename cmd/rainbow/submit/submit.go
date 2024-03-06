package submit

import (
	"context"
	"fmt"
	"log"
	"strings"

	js "github.com/compspec/jobspec-go/pkg/jobspec/v1"
	"github.com/converged-computing/rainbow/pkg/client"
	"github.com/converged-computing/rainbow/pkg/config"
)

// Run will check a manifest list of artifacts against a host machine
// For now, the host machine parameters will be provided as flags
func Run(
	host, jobName, command string,
	nodes, tasks int,
	token, clusterName,
	database, cfgFile string,
	selectionAlgorithm string,
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

	// Convert the simple command / nodes / etc into a JobSpec
	js, err := js.NewSimpleJobspec(jobName, command, int32(nodes), int32(tasks))
	if err != nil {
		return nil
	}

	// Read in the config, if provided, TODO we need a set of tokens here?
	cfg, err := config.NewRainbowClientConfig(cfgFile, "", "", database, selectionAlgorithm)
	if err != nil {
		return err
	}

	// The cluster name and token provided here are in reference to a cluster
	cfg.AddCluster(clusterName, token)

	// Submission is always with a configuration
	response, err := c.SubmitJob(context.Background(), js, cfg)
	if err != nil {
		return err
	}

	// If we get here, success! Dump all the stuff.
	//log.Printf("status: %s", response.Status)
	//log.Printf(" token: %s", response.Token)
	log.Println(response)
	return nil
}
