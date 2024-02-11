package main

import (
	"context"
	"flag"
	"log"

	"github.com/converged-computing/rainbow/pkg/client"
)

var (
	host        string
	clusterName string

	// set at build time
	version = "v0.0.1-default"
)

func main() {
	flag.StringVar(&host, "host", "localhost:50051", "Scheduler server address (host:port)")
	flag.StringVar(&clusterName, "cluster", "keebler", "Name of cluster to register")
	flag.Parse()

	log.Printf("creating client (%s)...", version)

	c, err := client.NewClient(host)
	if err != nil {
		log.Fatalf("error while creating client: %v", err)
	}

	log.Printf("registering cluster: %s", clusterName)

	// Last argument is secret, empty for now
	m, err := c.Register(context.Background(), clusterName, "")
	if err != nil {
		log.Fatalf("error while running client: %v", err)
	}
	log.Printf("received response: %s", m)
}
