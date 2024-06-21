package receive

import (
	"context"
	"log"

	"github.com/converged-computing/rainbow/pkg/client"
	"github.com/converged-computing/rainbow/pkg/config"
)

// Run will check a manifest list of artifacts against a host machine
// For now, the host machine parameters will be provided as flags
func Run(
	c client.Client,
	cluster, secret string,
	maxJobs int,
	cfgFile string,

) error {

	// Note that 0 or below indicates "show all jobs"
	if maxJobs >= 1 {
		log.Printf("receive jobs: %d", maxJobs)
	}

	// Read in the config, if provided, TODO we need a set of tokens here?
	cfg, err := config.NewRainbowClientConfig(cfgFile, cluster, secret, "", "", "")
	if err != nil {
		return err
	}

	// Last argument is secret, empty for now
	// Specific name and secret for the cluster asking for jobs
	response, err := c.ReceiveJobs(
		context.Background(),
		cfg.Cluster.Name,
		cfg.Cluster.Secret,
		int32(maxJobs),
	)
	if err != nil {
		return err
	}

	jobids := []int32{}
	log.Printf("🌀️ Received %d jobs!\n", len(response.Jobs))
	for jobid, jobstr := range response.Jobs {
		// TODO add load yaml function to jobspec go
		log.Printf("%d : %s", jobid, jobstr)
		jobids = append(jobids, jobid)
	}

	// Did we find jobs?
	if len(jobids) > 0 {
		log.Printf("✅️ Accepting %d jobs!\n", len(jobids))
		response, err := c.AcceptJobs(context.Background(), cfg.Cluster.Name, cfg.Cluster.Secret, jobids)
		if err != nil {
			return err
		}
		log.Printf("%s\n", response)
	}
	return nil
}
