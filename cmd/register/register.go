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
	secret      string
	version     = "v0.0.1-default"
)

func main() {
	flag.StringVar(&host, "host", "localhost:50051", "Scheduler server address (host:port)")
	flag.StringVar(&clusterName, "cluster", "keebler", "Name of cluster to register")
	flag.StringVar(&secret, "secret", "chocolate-cookies", "Registration 'secret'")
	flag.Parse()

	log.Printf("creating client (%s)...", version)

	c, err := client.NewClient(host)
	if err != nil {
		log.Fatalf("error while creating client: %v", err)
	}

	log.Printf("registering cluster: %s", clusterName)

	// Last argument is secret, empty for now
	response, err := c.Register(context.Background(), clusterName, secret)
	if err != nil {
		log.Fatalf("error while running client: %v", err)
	}

	// If we get here, success! Dump all the stuff.
	log.Printf("status: %s", response.Status)
	log.Printf(" token: %s", response.Token)
}
