package main

import (
	"fmt"
	"log"
	"os"

	"github.com/akamensky/argparse"
	"github.com/converged-computing/rainbow/cmd/rainbow/config"
	deleteCli "github.com/converged-computing/rainbow/cmd/rainbow/delete"
	"github.com/converged-computing/rainbow/cmd/rainbow/receive"
	"github.com/converged-computing/rainbow/cmd/rainbow/register"
	"github.com/converged-computing/rainbow/cmd/rainbow/submit"
	"github.com/converged-computing/rainbow/cmd/rainbow/update"
	"github.com/converged-computing/rainbow/pkg/certs"
	"github.com/converged-computing/rainbow/pkg/client"
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
	deleteCmd := parser.NewCommand("delete", "Delete a subsystem (including the cluster)")
	receiveCmd := parser.NewCommand("receive", "Receive and accept jobs")
	registerClusterCmd := registerCmd.NewCommand("cluster", "Register a new cluster")
	updateCmd := parser.NewCommand("update", "Update a cluster")

	// Configuration
	configCmd := parser.NewCommand("config", "Interact with rainbow configs")
	configInitCmd := configCmd.NewCommand("init", "Create a new configuration file")
	cfg := parser.String("", "config-path", &argparse.Options{Help: "Configuration file for cluster credentials"})

	// Credentials for client tls
	caCertFile := parser.String("", "ca-cert", &argparse.Options{Help: "Client CA cert file"})
	certFile := parser.String("", "cert", &argparse.Options{Help: "Client cert file"})
	keyFile := parser.String("", "key", &argparse.Options{Help: "Client key file"})

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

	// Delete arguments
	deleteTarget := deleteCmd.String("", "delete-subsystem", &argparse.Options{Default: "", Help: "Subsystem to delete, not provided defaults to cluster."})
	deleteSecret := deleteCmd.String("", "delete-secret", &argparse.Options{Default: "", Help: "Secret to authenticate delete"})

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

	// Generate certificate manager
	cert, err := certs.NewClientCertificate(*caCertFile, *certFile, *keyFile)
	if err != nil {
		log.Fatalf("error creating certificate manager: %v", err)
	}

	// Config is the only command that doesn't require the client
	if configCmd.Happened() && configInitCmd.Happened() {
		err := config.RunInit(*cfg, *clusterName, *selectAlgo, *matchAlgo)
		if err != nil {
			log.Fatalf("Issue with config: %s\n", err)
		}
		return
	}

	// Create the client to be used across calls
	client, err := client.NewClient(*host, cert)
	if err != nil {
		log.Fatalf("Issue creating client: %s\n", err)
	}

	if deleteCmd.Happened() {
		err := deleteCli.Run(
			client,
			*clusterName,
			*deleteTarget,
			*deleteSecret,
		)
		if err != nil {
			log.Fatalf("Issue with delete: %s\n", err)
		}

	} else if stateCmd.Happened() {
		err := update.UpdateState(
			client,
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
				client,
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
				client,
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
			client,
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
			client,
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
