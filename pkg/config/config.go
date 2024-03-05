package config

import (
	"encoding/json"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

var (
	defaultSelectionAlgorithm = "random"
	defaultGraphDatabase      = "memory"
)

// RainbowConfig is a static file that holds configuration parameteres
// for a client to access one or more clusters. We can eventually explore
// a logical grouping of clusters to have one access credential, but this
// works for now
type RainbowConfig struct {

	// Configuration section for Rainbow
	Scheduler RainbowScheduler `json:"scheduler"`

	// A "self" credential, saved on register / used to request (receive)
	Cluster ClusterCredential `json:"cluster,omitempty"`

	// Graph database selected
	GraphDatabase GraphDatabase `json:"graph"`

	// One or more clusters to submit to
	Clusters []ClusterCredential `json:"clusters"`
}

type RainbowScheduler struct {

	// Secret to register with the cluster
	// Absolutely should come from environment
	Secret    string             `json:"secret" yaml:"secret" envconfig:"RAINBOW_SECRET"`
	Name      string             `json:"name" yaml:"name" envconfig:"RAINBOW_SCHEDULER_NAME"`
	Algorithm SelectionAlgorithm `json:"algorithm" yaml:"algorithm"`
}

type SelectionAlgorithm struct {
	Name    string            `json:"name" yaml:"name" envconfig:"RAINBOW_SCHDULER_ALGORITHM"`
	Options map[string]string `json:"options,omitempty" yaml:"options,omitempty"`
}

// ClusterCredential holds a name and access token for a cluster
// When used for a "self" cluster, we have a name and secret
// When used for a "submit to" cluster, we have a name and token
type ClusterCredential struct {
	Name   string `json:"name" yaml:"name"`
	Token  string `json:"token,omitempty" yaml:"token,omitempty"`
	Secret string `json:"secret,omitempty" yaml:"secret,omitempty"`
}

// A Graph Database Backend takes a name and can handle options
type GraphDatabase struct {
	Name    string            `json:"name,omitempty" yaml:"name,omitempty"`
	Host    string            `json:"host,omitempty" yaml:"host,omitempty"`
	Options map[string]string `json:"options,omitempty" yaml:"options,omitempty"`
}

// ToYaml serializes to yaml
func (c *RainbowConfig) ToYaml() (string, error) {
	out, err := yaml.Marshal(c)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// ToJson serializes to json
func (c *RainbowConfig) ToJson() (string, error) {
	out, err := json.MarshalIndent(c, "", " ")
	if err != nil {
		return "", err
	}
	return string(out), nil
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
	database,
	selectionAlgorithm string,
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
	config.GraphDatabase.Name = defaultGraphDatabase
	if database != "" {
		config.GraphDatabase.Name = database
	}

	// Scheduling algorithm defaults to random selection
	algo := SelectionAlgorithm{Name: defaultSelectionAlgorithm, Options: map[string]string{}}
	config.Scheduler.Algorithm = algo
	if selectionAlgorithm == "" {
		config.Scheduler.Algorithm.Name = selectionAlgorithm
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
