package register

import (
	"context"
	"fmt"
	"log"

	"github.com/converged-computing/rainbow/pkg/client"
	"github.com/converged-computing/rainbow/pkg/config"
)

// RegisterSubsystem registers a subsystem
func RegisterSubsystem(
	host,
	clusterName,
	subsystemNodes,
	subsystem,
	cfgFile string,
) error {

	c, err := client.NewClient(host)
	if err != nil {
		return err
	}

	// A config file is required here
	if cfgFile == "" {
		return fmt.Errorf("an existing configuration file is required to register a subsystem")
	}
	if subsystem == "" {
		return fmt.Errorf("a subsystem name is required to register")
	}
	// Read in the config, if provided, command line takes preference
	cfg, err := config.NewRainbowClientConfig(cfgFile, "", "", "", "", "")
	if err != nil {
		return err
	}

	log.Printf("registering subsystem to cluster: %s", cfg.Scheduler.Name)

	// Last argument is subsystem name, which we can derive from graph
	response, err := c.RegisterSubsystem(
		context.Background(),
		cfg.Cluster.Name,
		cfg.Cluster.Secret,
		subsystemNodes,
		subsystem,
	)
	// If we get here, success! Dump all the stuff.
	log.Printf("%s", response)
	return err

}
