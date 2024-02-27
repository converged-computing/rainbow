package main

import (
	"context"
	"flag"
	"log"

	"github.com/converged-computing/rainbow/pkg/config"
	"github.com/converged-computing/rainbow/pkg/server"
	"github.com/converged-computing/rainbow/pkg/types"

	// Register database backends
	_ "github.com/converged-computing/rainbow/backends/memory"
)

var (
	host        string
	name        = "rainbow"
	sqliteFile  = "rainbow.db"
	configFile  = ""
	database    = ""
	cleanup     = false
	secret      = "chocolate-cookies"
	globalToken = ""
)

func main() {
	flag.StringVar(&host, "host", ":50051", "Server address (host:port)")
	flag.StringVar(&name, "name", name, "Server name (default: rainbow)")
	flag.StringVar(&sqliteFile, "db", sqliteFile, "sqlite3 database file (default: rainbow.db)")
	flag.StringVar(&globalToken, "global-token", name, "global token for cluster access (not recommended)")
	flag.StringVar(&secret, "secret", secret, "secret to validate registration (default: chocolate-cookies)")
	flag.StringVar(&database, "graph-database", database, "graph database backend (defaults to memory)")
	flag.StringVar(&configFile, "config", configFile, "rainbow config file")
	flag.BoolVar(&cleanup, "cleanup", cleanup, "cleanup previous sqlite database (default: false)")
	flag.Parse()

	// Load (or generate a default)  config file here, if provided
	cfg, err := config.NewRainbowClientConfig(configFile, name, secret, database)
	if err != nil {
		log.Fatalf("error while creating server: %v", err)
	}

	// create server
	log.Print("creating 🌈️ server...")
	s, err := server.NewServer(cfg, types.Version, sqliteFile, cleanup, globalToken)
	if err != nil {
		log.Fatalf("error while creating server: %v", err)
	}
	defer s.Stop()

	// Give a warning if the globalToken is set
	if globalToken != "" {
		log.Printf("⚠️ WARNING: global-token is set, use with caution.")
	}

	// run server
	log.Printf("starting scheduler server: %s", s.String())
	if err := s.Start(context.Background(), host); err != nil {
		log.Fatalf("error while running scheduler server: %v", err)
	}
	log.Printf("🌈️ done 🌈️")
}
