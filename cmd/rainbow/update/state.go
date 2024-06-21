package update

import (
	"context"
	"fmt"
	"log"

	"github.com/converged-computing/rainbow/pkg/client"
	"github.com/converged-computing/rainbow/pkg/config"
)

// UpdateState updates state for a cluster
func UpdateState(
	c client.Client,
	clusterName,
	stateFile,
	cfgFile string,
) error {

	// A config file is required here
	if cfgFile == "" {
		return fmt.Errorf("an existing configuration file is required to update an existing cluster")
	}
	if stateFile == "" {
		return fmt.Errorf("a state file (json with key value pairs) is required to update state")
	}
	// Read in the config, if provided, command line takes preference
	cfg, err := config.NewRainbowClientConfig(cfgFile, "", "", "", "", "")
	if err != nil {
		return err
	}

	log.Printf("updating state for cluster: %s", cfg.Scheduler.Name)

	// Last argument is subsystem name, which we can derive from graph
	response, err := c.UpdateState(
		context.Background(),
		cfg.Cluster.Name,
		cfg.Cluster.Secret,
		stateFile,
	)
	// If we get here, success! Dump all the stuff.
	log.Printf("%s", response)
	return err

}
