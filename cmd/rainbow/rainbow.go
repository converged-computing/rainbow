package main

import (
	"fmt"
	"log"
	"os"

	"github.com/akamensky/argparse"
	config "github.com/converged-computing/rainbow/cmd/rainbow/config"
	register "github.com/converged-computing/rainbow/cmd/rainbow/register"
	request "github.com/converged-computing/rainbow/cmd/rainbow/request"
	submit "github.com/converged-computing/rainbow/cmd/rainbow/submit"
	"github.com/converged-computing/rainbow/pkg/types"
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
	requestCmd := parser.NewCommand("request", "Request to inspect some max jobs assigned to a cluster")

	// Configuration
	configCmd := parser.NewCommand("config", "Interact with rainbow configs")
	configInitCmd := configCmd.NewCommand("init", "Create a new configuration file")
	configPath := parser.String("", "config-path", &argparse.Options{Default: "rainbow-config.yaml", Help: "Rainbow config file"})

	// Shared values
	host := parser.String("", "host", &argparse.Options{Default: "localhost:50051", Help: "Scheduler server address (host:port)"})
	clusterName := parser.String("", "cluster-name", &argparse.Options{Help: "Name of cluster to register"})
	cfg := parser.String("", "config", &argparse.Options{Help: "Configuration file for cluster credentials"})
	graphDatabase := parser.String("", "graph-database", &argparse.Options{Help: "Graph database backend to use"})

	// Request Jobs
	clusterSecret := requestCmd.String("", "request-secret", &argparse.Options{Help: "Cluster 'secret' to retrieve jobs"})
	maxJobs := requestCmd.Int("j", "max-jobs", &argparse.Options{Help: "Maximum number of jobs to request"})
	acceptJobs := requestCmd.Int("", "accept-jobs", &argparse.Options{Default: 0, Help: "Jobs to accept from the set"})

	// Register
	secret := registerCmd.String("s", "secret", &argparse.Options{Default: defaultSecret, Help: "Registration 'secret'"})
	clusterNodes := registerCmd.String("", "cluster-nodes", &argparse.Options{Help: "Cluster nodes json (JGF v2)"})

	// Submit (note that command for now needs to be in quotes to get the whole thing)
	token := submitCmd.String("", "token", &argparse.Options{Default: defaultSecret, Help: "Client token to submit jobs with."})
	nodes := submitCmd.Int("n", "nodes", &argparse.Options{Default: 1, Help: "Number of nodes to request"})
	tasks := submitCmd.Int("t", "tasks", &argparse.Options{Help: "Number of tasks to request (per node? total?)"})
	command := submitCmd.String("c", "command", &argparse.Options{Default: defaultSecret, Help: "Command to submit"})
	jobName := submitCmd.String("", "job-name", &argparse.Options{Help: "Name for the job (defaults to first command)"})

	// Now parse the arguments
	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Println(Header)
		fmt.Println(parser.Usage(err))
		return
	}

	if configCmd.Happened() && configInitCmd.Happened() {
		err := config.RunInit(*configPath)
		if err != nil {
			log.Fatalf("Issue with config: %s\n", err)
		}

	} else if registerCmd.Happened() {
		err := register.Run(*host, *clusterName, *clusterNodes, *secret, *cfg, *graphDatabase)
		if err != nil {
			log.Fatalf("Issue with register: %s\n", err)
		}
	} else if requestCmd.Happened() {
		err := request.Run(*host, *clusterName, *clusterSecret, *maxJobs, *acceptJobs, *cfg)
		if err != nil {
			log.Fatalf("Issue with request jobs: %s\n", err)
		}
	} else if submitCmd.Happened() {
		err := submit.Run(*host, *jobName, *command, *nodes, *tasks, *token, *clusterName, *cfg)
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
