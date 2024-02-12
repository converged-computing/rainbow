package main

import (
	"fmt"
	"log"
	"os"

	"github.com/akamensky/argparse"
	register "github.com/converged-computing/rainbow/cmd/rainbow/register"
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

	parser := argparse.NewParser("rainbow", "Interact with a rainbow multi-cluster")
	versionCmd := parser.NewCommand("version", "See the version of compspec")
	registerCmd := parser.NewCommand("register", "Register a new cluster")
	submitCmd := parser.NewCommand("submit", "Submit a job to a rainbow cluster")

	// Shared values
	host := parser.String("", "host", &argparse.Options{Default: "localhost:50051", Help: "Scheduler server address (host:port)"})
	clusterName := parser.String("", "cluster-name", &argparse.Options{Default: "keebler", Help: "Name of cluster to register"})

	// Register
	secret := registerCmd.String("s", "secret", &argparse.Options{Default: defaultSecret, Help: "Registration 'secret'"})

	// Submit (note that command for now needs to be in quotes to get the whole thing)
	submitSecret := submitCmd.String("s", "secret", &argparse.Options{Default: defaultSecret, Help: "Registration 'secret'"})
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

	if registerCmd.Happened() {
		err := register.Run(*host, *clusterName, *secret)
		if err != nil {
			log.Fatalf("Issue with register: %s\n", err)
		}
	} else if submitCmd.Happened() {
		err := submit.Run(*host, *jobName, *command, *nodes, *tasks, *submitSecret, *clusterName)
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
