package delete

import (
	"context"
	"log"

	"github.com/converged-computing/rainbow/pkg/client"
)

func Run(
	c client.Client,
	clusterName, subsystem, secret string,
) error {

	// Submission is always with a configuration
	if subsystem == "cluster" || subsystem == "" {
		response, err := c.Delete(context.Background(), clusterName, secret, subsystem)
		if err != nil {
			return err
		}
		log.Println(response)
		return nil
	}
	response, err := c.DeleteSubsystem(context.Background(), clusterName, secret, subsystem)
	if err != nil {
		return err
	}
	log.Println(response)
	return nil
}
