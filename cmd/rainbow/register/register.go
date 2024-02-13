package extract

import (
	"context"
	"log"

	"github.com/converged-computing/rainbow/pkg/client"
)

// Run will run an extraction of host metadata
func Run(host, clusterName, secret string) error {
	c, err := client.NewClient(host)
	if err != nil {
		return err
	}

	log.Printf("registering cluster: %s", clusterName)

	// Last argument is secret, empty for now
	response, err := c.Register(context.Background(), clusterName, secret)
	if err != nil {
		return err
	}

	// If we get here, success! Dump all the stuff.
	log.Printf("status: %s", response.Status)
	log.Printf("secret: %s", response.Secret)
	log.Printf(" token: %s", response.Token)
	return nil
}
