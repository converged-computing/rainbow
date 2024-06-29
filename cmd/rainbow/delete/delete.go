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
		_, err := c.Delete(context.Background(), clusterName, secret, subsystem)
		if err != nil {
			return err
		}
		log.Printf("🔥️ Cluster %s has been deleted.\n", clusterName)
		return nil
	}
	_, err := c.DeleteSubsystem(context.Background(), clusterName, secret, subsystem)
	if err != nil {
		return err
	}
	log.Printf("🔥️ Cluster %s subsystem %s has been deleted.\n", clusterName, subsystem)
	return nil
}
