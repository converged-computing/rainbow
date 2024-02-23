package match

import (
	"context"
	"log"

	"github.com/converged-computing/rainbow/pkg/client"
	"github.com/converged-computing/rainbow/pkg/utils"
)

// Run will check a manifest list of artifacts against a host machine
// For now, the host machine parameters will be provided as flags
func Run(
	host, cluster, secret string,
	maxJobs int,
	acceptJobs int,
	cfgFile string,
) error {

	c, err := client.NewClient(host)
	if err != nil {
		return nil
	}

	// Note that 0 or below indicates "show all jobs"
	if maxJobs >= 1 {
		log.Printf("request jobs: %d", maxJobs)
	}

	// TODO this needs to handle submission to multiple clusters
	// Read in the config, if provided, command line takes preference
	//cfg, err := config.NewRainbowClientConfig(cfgFile, clusterName, secret)
	//if err != nil {
	//	return err
	//}

	// Last argument is secret, empty for now
	response, err := c.RequestJobs(context.Background(), cluster, secret, int32(maxJobs))
	if err != nil {
		return err
	}

	jobids := []int32{}
	log.Printf("🌀️ Found %d jobs!\n", len(response.Jobs))
	for jobid, jobstr := range response.Jobs {
		log.Printf("%d : %s", jobid, jobstr)
		jobids = append(jobids, jobid)
	}

	// We can only accept the max number we get back
	if acceptJobs > len(response.Jobs) {
		acceptJobs = len(response.Jobs)
	}

	// Are we accepting jobs?
	if acceptJobs > 0 {

		log.Printf("✅️ Accepting %d jobs!\n", acceptJobs)
		shuffled := utils.ShuffleJobs(jobids)

		// Randomly select for the example
		accepted := shuffled[0:acceptJobs]
		for _, jobid := range accepted {
			log.Printf("   %d", jobid)
		}
		response, err := c.AcceptJobs(context.Background(), cluster, secret, accepted)
		if err != nil {
			return err
		}
		log.Printf("%s\n", response)
	}
	return nil
}
