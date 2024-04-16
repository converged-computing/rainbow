package register

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/converged-computing/rainbow/pkg/client"
	"github.com/converged-computing/rainbow/pkg/config"
)

// Run will register the cluster with rainbow
func Run(
	c client.Client,
	clusterName,
	clusterNodes,
	secret string,
	saveSecret bool,
	cfgFile,
	graphDatabase,
	subsystem,
	selectionAlgorithm string,
	matchAlgorithm string,

) error {
	if clusterName == "" {
		return fmt.Errorf("s --cluster-name is required")
	}
	// Read in the config, if provided, command line takes preference
	cfg, err := config.NewRainbowClientConfig(
		cfgFile,
		clusterName,
		secret,
		graphDatabase,
		selectionAlgorithm,
		matchAlgorithm,
	)
	if err != nil {
		return err
	}

	log.Printf("registering cluster: %s", cfg.Scheduler.Name)

	// Last argument is secret, empty for now
	response, err := c.Register(
		context.Background(),
		cfg.Scheduler.Name,
		cfg.Scheduler.Secret,
		clusterNodes,
		subsystem,
	)
	if err != nil {
		return err
	}

	// If we get here, success! Dump all the stuff.
	log.Printf("status: %s", response.Status)
	log.Printf("secret: %s", response.Secret)
	log.Printf(" token: %s", response.Token)

	// If we have a config file and flag is provided to save secret, do it.
	if saveSecret && cfgFile != "" {
		log.Printf("Saving cluster secret to %s\n", cfgFile)
		cfg.Cluster = config.ClusterCredential{Secret: response.Secret, Name: clusterName}

		// Assume we want to submit to our cluster too
		newCluster := config.ClusterCredential{Token: response.Token, Name: clusterName}
		cfg.Clusters = []config.ClusterCredential{newCluster}
		yaml, err := cfg.ToYaml()
		if err != nil {
			return err
		}
		err = os.WriteFile(cfgFile, []byte(yaml), 0644)
		if err != nil {
			return err
		}
	}
	return nil
}
