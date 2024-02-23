package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// RainbowConfig is a static file that holds configuration parameteres
// for a client to access one or more clusters. We can eventually explore
// a logical grouping of clusters to have one access credential, but this
// works for now
type RainbowConfig struct {

	// Configuration section for Rainbow
	Scheduler RainbowScheduler `json:"scheduler"`

	// One or more clusters to submit to
	Clusters []ClusterCredential `json:"clusters"`
}

type RainbowScheduler struct {

	// Secret to register with the cluster
	// Absolutely should come from environment
	Secret string `json:"secret" yaml:"secret" envconfig:"RAINBOW_SECRET"`
	Name   string `json:"name" yaml:"name" envconfig:"RAINBOW_CLUSTER_NAME"`
}

// ClusterCredential holds a name and access token for a cluster
type ClusterCredential struct {
	Name  string `json:"name" yaml:"name"`
	Token string `json:"token" yaml:"token"`
}

// NewRainbowClientConfig reads in a config or generates a new one
func NewRainbowClientConfig(filename, clusterName, secret string) (*RainbowConfig, error) {

	config := RainbowConfig{}
	var err error

	// If we are given a filename, load it
	if filename != "" {
		err = config.Load(filename)
		if err != nil {
			return &config, err
		}
	}

	// Command line takes precedence
	if clusterName != "" {
		config.Scheduler.Name = clusterName
	}
	if secret != "" {
		config.Scheduler.Secret = secret
	}
	return &config, err
}

// Load a filename into the rainbow config
func (cfg *RainbowConfig) Load(filename string) error {

	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&cfg)
	if err != nil {
		return err
	}
	return nil
}
