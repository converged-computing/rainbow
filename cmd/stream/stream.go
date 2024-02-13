package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/converged-computing/rainbow/pkg/client"
	"github.com/converged-computing/rainbow/pkg/types"
)

var (
	target  string
	timeout float64

	// set at build time
	version = "v0.0.1-default"
)

func main() {
	flag.StringVar(&target, "target", "localhost:50051", "Server address (host:port)")
	flag.Float64Var(&timeout, "timeout", 5, "Timeout in seconds (default: 5)")
	flag.Parse()

	log.Printf("creating client (%s)...", version)

	c, err := client.NewClient(target)
	if err != nil {
		log.Fatalf("error while creating client: %v", err)
	}

	// Contact the server and print out its response.
	d := time.Duration(float64(time.Second) * timeout)
	ctx, cancel := context.WithTimeout(context.Background(), d)
	defer cancel()

	log.Printf("starting stream with %v timeout...", d)
	if err := c.Stream(ctx, types.MockedMessageProvider); err != nil {
		log.Fatalf("error: %v", err)
	}

	log.Printf("done")
}
