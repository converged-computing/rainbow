package config

import (
	"fmt"
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

	// Graph database selected
	GraphDatabase GraphDatabase `json:"graph"`

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

// A Graph Database Backend takes a name and can handle options
type GraphDatabase struct {
	Name    string            `json:"name,omitempty" yaml:"name,omitempty"`
	Host    string            `json:"host,omitempty" yaml:"host,omitempty"`
	Options map[string]string `json:"options,omitempty" yaml:"options,omitempty"`
}

// AddCluster adds a cluster on the fly to a config, likely for a one-off submit
func (c *RainbowConfig) AddCluster(clusterName, token string) error {

	if clusterName == "" {
		return fmt.Errorf("a cluster name is required")
	}
	if token == "" {
		return fmt.Errorf("a token for cluster %s is required", clusterName)
	}

	// Ensure we don't have it already
	for _, item := range c.Clusters {
		if item.Name == clusterName {
			return fmt.Errorf("cluster %s is already added to the configuration file", clusterName)
		}
	}
	newCluster := ClusterCredential{Name: clusterName, Token: token}
	c.Clusters = append(c.Clusters, newCluster)
	return nil
}

// NewRainbowClientConfig reads in a config or generates a new one
func NewRainbowClientConfig(
	filename,
	clusterName,
	secret,
	database string,
) (*RainbowConfig, error) {

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

	// By default we use the in-memory (vanilla, simple) database
	config.GraphDatabase.Name = "memory"
	if database != "" {
		config.GraphDatabase.Name = database
	}

	// Default host, for now is always this
	if config.GraphDatabase.Host == "" {
		config.GraphDatabase.Host = "127.0.0.1:50051"
	}
	return &config, err
}

// NewRainbowServerConfig creates a default empty config for a server
func NewRainbowServerConfig(name string) *RainbowConfig {
	config := RainbowConfig{Scheduler: RainbowScheduler{Name: name}}
	config.Clusters = make([]ClusterCredential, 0)
	return &config
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
