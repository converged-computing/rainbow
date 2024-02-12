package main

import (
	"context"
	"flag"
	"log"

	"github.com/converged-computing/rainbow/pkg/server"
)

var (
	address     string
	name        = "rainbow"
	sqliteFile  = "rainbow.db"
	environment = "development"

	// Remove the previous database
	skipCleanup = false
	secret      = "chocolate-cookies"
	version     = "v0.0.1-default"
)

func main() {
	flag.StringVar(&address, "address", ":50051", "Server address (host:port)")
	flag.StringVar(&name, "name", name, "Server name (default: rainbow)")
	flag.StringVar(&sqliteFile, "db", sqliteFile, "sqlite3 database file (default: rainbow.db)")
	flag.StringVar(&secret, "secret", secret, "secret to validate registration (default: chocolate-cookies)")
	flag.StringVar(&environment, "environment", environment, "environment (default: development)")
	flag.BoolVar(&skipCleanup, "skip-cleanup", skipCleanup, "skip cleanup of previous sqlite database (default: false)")
	flag.Parse()

	// create server
	log.Print("creating 🌈️ server...")
	s, err := server.NewServer(name, version, environment, sqliteFile, !skipCleanup, secret)
	if err != nil {
		log.Fatalf("error while creating server: %v", err)
	}
	defer s.Stop()

	// run server
	log.Printf("starting scheduler server: %s", s.String())
	if err := s.Start(context.Background(), address); err != nil {
		log.Fatalf("error while running scheduler server: %v", err)
	}
	log.Printf("🌈️ done 🌈️")
}
