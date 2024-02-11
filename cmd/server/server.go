package main

import (
	"context"
	"flag"
	"log"

	"github.com/converged-computing/rainbow/pkg/server"
)

var (
	address     string
	name        = "server"
	environment = "development"

	// set at build time
	version = "v0.0.1-default"
)

func main() {
	flag.StringVar(&address, "address", ":50051", "Server address (host:port)")
	flag.StringVar(&name, "name", name, "Server name (default: server)")
	flag.StringVar(&environment, "environment", environment, "Server environment (default: development)")
	flag.Parse()

	// create server
	log.Print("creating server...")
	s, err := server.NewServer(name, version, environment)
	if err != nil {
		log.Fatalf("error while creating server: %v", err)
	}
	defer s.Stop()

	// run server
	log.Printf("starting server: %s", s.String())
	if err := s.Start(context.Background(), address); err != nil {
		log.Fatalf("error while running server: %v", err)
	}

	log.Printf("done")
}
