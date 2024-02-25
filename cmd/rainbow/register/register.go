package extract

import (
	"context"
	"log"

	"github.com/converged-computing/rainbow/pkg/client"
	"github.com/converged-computing/rainbow/pkg/config"
)

// Run will register the cluster with rainbow
func Run(host, clusterName, clusterNodes, secret, cfgFile string) error {
	c, err := client.NewClient(host)
	if err != nil {
		return err
	}

	// Read in the config, if provided, command line takes preference
	cfg, err := config.NewRainbowClientConfig(cfgFile, clusterName, secret)
	if err != nil {
		return err
	}

	log.Printf("registering cluster: %s", cfg.Scheduler.Name)

	// Last argument is secret, empty for now
	response, err := c.Register(context.Background(), cfg.Scheduler.Name, cfg.Scheduler.Secret, clusterNodes)
	if err != nil {
		return err
	}

	// If we get here, success! Dump all the stuff.
	log.Printf("status: %s", response.Status)
	log.Printf("secret: %s", response.Secret)
	log.Printf(" token: %s", response.Token)
	return nil
}
