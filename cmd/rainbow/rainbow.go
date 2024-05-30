package main

import (
	"fmt"
	"log"
	"os"

	"github.com/akamensky/argparse"
	"github.com/converged-computing/rainbow/cmd/rainbow/config"
	"github.com/converged-computing/rainbow/cmd/rainbow/receive"
	"github.com/converged-computing/rainbow/cmd/rainbow/register"
	"github.com/converged-computing/rainbow/cmd/rainbow/submit"
	"github.com/converged-computing/rainbow/cmd/rainbow/update"
	"github.com/converged-computing/rainbow/pkg/types"

	// Register database backends and selection algorithms
	_ "github.com/converged-computing/rainbow/plugins/algorithms/match"
	_ "github.com/converged-computing/rainbow/plugins/backends/memgraph"
	_ "github.com/converged-computing/rainbow/plugins/backends/memory"
	_ "github.com/converged-computing/rainbow/plugins/backends/neo4j"
	_ "github.com/converged-computing/rainbow/plugins/selection/constraint"
	_ "github.com/converged-computing/rainbow/plugins/selection/random"
)

var (
	Header = `
    •  ┓
┏┓┏┓┓┏┓┣┓┏┓┓┏┏
┛ ┗┻┗┛┗┗┛┗┛┗┻┛
`

	defaultSecret = "chocolate-cookies"
)

func RunVersion() {
	fmt.Printf("🌈️ rainbow version %s\n", types.Version)
}

func main() {

	parser := argparse.NewParser("rainbow", "Interact with a rainbow scheduler")
	versionCmd := parser.NewCommand("version", "See the version of rainbow")
	registerCmd := parser.NewCommand("register", "Register a new cluster")
	submitCmd := parser.NewCommand("submit", "Submit a job to a rainbow scheduler")
	receiveCmd := parser.NewCommand("receive", "Receive and accept jobs")
	registerClusterCmd := registerCmd.NewCommand("cluster", "Register a new cluster")
	updateCmd := parser.NewCommand("update", "Update a cluster")

	// Configuration
	configCmd := parser.NewCommand("config", "Interact with rainbow configs")
	configInitCmd := configCmd.NewCommand("init", "Create a new configuration file")
	cfg := parser.String("", "config-path", &argparse.Options{Help: "Configuration file for cluster credentials"})

	// Shared values
	host := parser.String("", "host", &argparse.Options{Default: "localhost:50051", Help: "Scheduler server address (host:port)"})
	clusterName := parser.String("", "cluster-name", &argparse.Options{Help: "Name of cluster to register"})
	graphDatabase := parser.String("", "graph-database", &argparse.Options{Help: "Graph database backend to use"})
	selectAlgo := parser.String("", "select-algorithm", &argparse.Options{Default: "random", Help: "Selection algorithm for final cluster selection (defaults to random)"})
	matchAlgo := parser.String("", "match-algorithm", &argparse.Options{Default: "match", Help: "Match algorithm for graph database (defaults to match)"})

	// Receive Jobs
	clusterSecret := receiveCmd.String("", "request-secret", &argparse.Options{Help: "Cluster 'secret' to retrieve jobs"})
	maxJobs := receiveCmd.Int("j", "max-jobs", &argparse.Options{Help: "Maximum number of jobs to accept"})

	// Register Shared arguments
	clusterNodes := registerCmd.String("", "nodes-json", &argparse.Options{Help: "Cluster nodes json (JGF v2)"})

	// Cluster register arguments
	secret := registerClusterCmd.String("s", "secret", &argparse.Options{Default: defaultSecret, Help: "Registration 'secret'"})
	subsystem := registerCmd.String("", "subsystem", &argparse.Options{Help: "Subsystem to register cluster to (defaults to dominant, nodes)"})
	saveSecret := registerClusterCmd.Flag("", "save", &argparse.Options{Help: "Save cluster secret to config file, if provided"})

	// Register subsystem (requires config file for authentication)
	subsysCmd := registerCmd.NewCommand("subsystem", "Register a new subsystem")

	// Update subcommands - currently just supported are state
	stateCmd := updateCmd.NewCommand("state", "Update the state for a known cluster")
	stateFile := stateCmd.String("", "state-file", &argparse.Options{Help: "JSON file with key, value attributes for the cluster"})

	// Submit (note that command for now needs to be in quotes to get the whole thing)
	token := submitCmd.String("", "token", &argparse.Options{Default: defaultSecret, Help: "Client token to submit jobs with."})
	nodes := submitCmd.Int("n", "nodes", &argparse.Options{Default: 1, Help: "Number of nodes to request"})
	tasks := submitCmd.Int("t", "tasks", &argparse.Options{Help: "Number of tasks to request (per node? total?)"})
	command := submitCmd.String("c", "command", &argparse.Options{Default: defaultSecret, Help: "Command to submit"})
	jobName := submitCmd.String("", "job-name", &argparse.Options{Help: "Name for the job (defaults to first command)"})
	jobspec := submitCmd.String("", "jobspec", &argparse.Options{Help: "A yaml Jobspec to submit"})

	// Now parse the arguments
	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Println(Header)
		fmt.Println(parser.Usage(err))
		return
	}

	if configCmd.Happened() && configInitCmd.Happened() {
		err := config.RunInit(*cfg, *clusterName, *selectAlgo, *matchAlgo)
		if err != nil {
			log.Fatalf("Issue with config: %s\n", err)
		}

	} else if stateCmd.Happened() {
		err := update.UpdateState(
			*host,
			*clusterName,
			*stateFile,
			*cfg,
		)
		if err != nil {
			log.Fatalf("Issue with register subsystem: %s\n", err)
		}

	} else if registerCmd.Happened() {

		if subsysCmd.Happened() {
			err := register.RegisterSubsystem(
				*host,
				*clusterName,
				*clusterNodes,
				*subsystem,
				*cfg,
			)
			if err != nil {
				log.Fatalf("Issue with register subsystem: %s\n", err)
			}
		} else if registerClusterCmd.Happened() {
			err := register.Run(
				*host,
				*clusterName,
				*clusterNodes,
				*secret,
				*saveSecret,
				*cfg,
				*graphDatabase,
				*subsystem,
				*selectAlgo,
				*matchAlgo,
			)
			if err != nil {
				log.Fatalf("Issue with register: %s\n", err)
			}
		} else {
			log.Fatal("Register requires a command.")
		}

	} else if receiveCmd.Happened() {
		err := receive.Run(
			*host,
			*clusterName,
			*clusterSecret,
			*maxJobs,
			*cfg,
		)
		if err != nil {
			log.Fatalf("Issue with request jobs: %s\n", err)
		}
	} else if submitCmd.Happened() {
		err := submit.Run(
			*host,
			*jobName,
			*command,
			*nodes,
			*tasks,
			*token,
			*jobspec,
			*clusterName,
			*graphDatabase,
			*cfg,
			*selectAlgo,
			*matchAlgo,
		)
		if err != nil {
			log.Fatal(err.Error())
		}
	} else if versionCmd.Happened() {
		RunVersion()
	} else {
		fmt.Println(Header)
		fmt.Println(parser.Usage(nil))
	}
}
