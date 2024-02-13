package match

import (
	"context"
	"fmt"
	"log"

	"github.com/converged-computing/rainbow/pkg/client"
)

// Run will check a manifest list of artifacts against a host machine
// For now, the host machine parameters will be provided as flags
func Run(
	host, cluster, secret string,
	maxJobs int,
) error {

	c, err := client.NewClient(host)
	if err != nil {
		return nil
	}

	// Further validation of job happens with client below
	if maxJobs < 1 {
		return fmt.Errorf(">1 job must be requested")
	}
	log.Printf("request jobs: %d", maxJobs)

	// Last argument is secret, empty for now
	response, err := c.RequestJobs(context.Background(), cluster, secret, int32(maxJobs))
	if err != nil {
		return err
	}
	log.Printf("🌀️ Found %d jobs!\n", len(response.Jobs))
	for jobid, jobstr := range response.Jobs {
		log.Printf("%d : %s", jobid, jobstr)
	}
	return nil
}
