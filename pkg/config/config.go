package config

import (
	"encoding/json"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

var (
	DefaultSelectionAlgorithm = "random"
	DefaultMatchAlgorithm     = "match"
	DefaultGraphDatabase      = "memory"
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
	Secret     string     `json:"secret" yaml:"secret" envconfig:"RAINBOW_SECRET"`
	Name       string     `json:"name" yaml:"name" envconfig:"RAINBOW_SCHEDULER_NAME"`
	Algorithms Algorithms `json:"algorithms" yaml:"algorithms"`
}

type Algorithms struct {
	Selection Algorithm `json:"selection" yaml:"selection"`
	Match     Algorithm `json:"match" yaml:"match"`
}

type Algorithm struct {
	Name    string            `json:"name" yaml:"name,omitempty"`
	Options map[string]string `json:"options,omitempty" yaml:"options,omitempty"`
}

// ClusterCredential holds a name and access token for a cluster
// When used for a "self" cluster, we have a name and secret
// When used for a "submit to" cluster, we have a name and token
type ClusterCredential struct {
	Name   string `json:"name,omitempty" yaml:"name,omitempty"`
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

// setAlgorithm sets the algorithms for the rainbow scheduler
// But only if they aren't already set!
func (c *RainbowConfig) setAlgorithms(selectAlgo, matchAlgo string) {
	sAlgo := Algorithm{Name: DefaultSelectionAlgorithm, Options: map[string]string{}}
	mAlgo := Algorithm{Name: DefaultMatchAlgorithm, Options: map[string]string{}}

	// Only set if our config is missing it
	if matchAlgo != "" && c.Scheduler.Algorithms.Match.Name == "" {
		mAlgo.Name = matchAlgo
		c.Scheduler.Algorithms.Match = mAlgo
	}
	if selectAlgo != "" && c.Scheduler.Algorithms.Selection.Name == "" {
		sAlgo.Name = selectAlgo
		c.Scheduler.Algorithms.Selection = sAlgo
	}
}

// ToJson serializes to json
func (c *RainbowConfig) ToJson() (string, error) {
	out, err := json.MarshalIndent(c, "", " ")
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// GetCluster returns a cluster, if it is known to the config
func (c *RainbowConfig) GetClusterToken(clusterName string) string {
	for _, c := range c.Clusters {
		if c.Name == clusterName {
			return c.Token
		}
	}
	return ""
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
	matchAlgorithm string,
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
	if config.GraphDatabase.Name == "" {
		config.GraphDatabase.Name = DefaultGraphDatabase
	}
	if database != "" {
		config.GraphDatabase.Name = database
	}

	// Scheduling algorithm defaults to random selection
	config.setAlgorithms(selectionAlgorithm, matchAlgorithm)

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
	f, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	return yaml.Unmarshal([]byte(f), cfg)
}
