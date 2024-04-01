package main

import (
	"context"
	"flag"
	"log"

	"github.com/converged-computing/rainbow/pkg/config"
	rlog "github.com/converged-computing/rainbow/pkg/logger"
	"github.com/converged-computing/rainbow/pkg/server"
	"github.com/converged-computing/rainbow/pkg/types"

	// Register database backends
	_ "github.com/converged-computing/rainbow/plugins/algorithms/match"
	_ "github.com/converged-computing/rainbow/plugins/algorithms/range"
	_ "github.com/converged-computing/rainbow/plugins/backends/memory"
	_ "github.com/converged-computing/rainbow/plugins/selection/random"
)

var (
	host string

	// default logging level of warning (none, info, warning)
	loggingLevel = 3
	name         = "rainbow"
	sqliteFile   = "rainbow.db"
	configFile   = ""
	matchAlgo    = "match"
	selectAlgo   = "random"
	database     = ""
	cleanup      = false
	secret       = "chocolate-cookies"
	globalToken  = ""
)

func main() {
	flag.StringVar(&host, "host", ":50051", "Server address (host:port)")
	flag.StringVar(&name, "name", name, "Server name (default: rainbow)")
	flag.StringVar(&sqliteFile, "db", sqliteFile, "sqlite3 database file (default: rainbow.db)")
	flag.StringVar(&globalToken, "global-token", name, "global token for cluster access (not recommended)")
	flag.StringVar(&secret, "secret", secret, "secret to validate registration (default: chocolate-cookies)")
	flag.StringVar(&database, "graph-database", database, "graph database backend (defaults to memory)")
	flag.StringVar(&selectAlgo, "select-algorithm", selectAlgo, "selection algorithm for final cluster selection (defaults to random)")
	flag.StringVar(&matchAlgo, "match-algorithm", matchAlgo, "match algorithm for graph (defaults to random)")
	flag.StringVar(&configFile, "config", configFile, "rainbow config file")
	flag.IntVar(&loggingLevel, "loglevel", loggingLevel, "rainbow logging level (0 to 5)")
	flag.BoolVar(&cleanup, "cleanup", cleanup, "cleanup previous sqlite database (default: false)")
	flag.Parse()

	// If the logging level isn't the default, set it
	if loggingLevel != rlog.DefaultLevel {
		rlog.SetLevel(loggingLevel)
	}

	// Load (or generate a default)  config file here, if provided
	cfg, err := config.NewRainbowClientConfig(configFile, name, secret, database, selectAlgo, matchAlgo)
	if err != nil {
		log.Fatalf("error while creating server: %v", err)
	}

	// create server
	log.Print("creating 🌈️ server...")
	s, err := server.NewServer(cfg, types.Version, sqliteFile, cleanup, globalToken, host)
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
