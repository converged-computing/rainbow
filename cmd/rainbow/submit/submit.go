package submit

import (
	"context"
	"fmt"
	"log"

	js "github.com/compspec/jobspec-go/pkg/nextgen/v1"
	"github.com/converged-computing/rainbow/pkg/client"
	"github.com/converged-computing/rainbow/pkg/config"
	jscli "github.com/converged-computing/rainbow/pkg/jobspec"
)

// Run will check a manifest list of artifacts against a host machine
// For now, the host machine parameters will be provided as flags
func Run(
	c client.Client,
	jobName, command string,
	nodes, tasks int,
	token, jobspec, clusterName,
	database, cfgFile string,
	selectAlgo, matchAlgo string,
) error {

	var err error
	jspec := &js.Jobspec{}
	if jobspec == "" {
		jspec, err = jscli.JobspecFromCommand(command, jobName, int32(nodes), int32(tasks))
		if err != nil {
			return err
		}
	} else {
		jspec, err = js.LoadJobspecYaml(jobspec)
		if err != nil {
			return err
		}

		// Validate the jobspec
		valid, err := jspec.Validate()
		if err != nil {
			return err
		}
		if !valid {
			return fmt.Errorf("jobspec is not valid")
		}
	}

	// Read in the config, if provided, TODO we need a set of tokens here?
	cfg, err := config.NewRainbowClientConfig(cfgFile, "", "", database, selectAlgo, matchAlgo)
	if err != nil {
		return err
	}

	// The cluster name and token provided here are in reference to a cluster
	cfg.AddCluster(clusterName, token)

	// Submission is always with a configuration
	response, err := c.SubmitJob(context.Background(), jspec, cfg)
	if err != nil {
		return err
	}

	// If we get here, success! Dump all the stuff.
	//log.Printf("status: %s", response.Status)
	//log.Printf(" token: %s", response.Token)
	log.Println(response)
	return nil
}
