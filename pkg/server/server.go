package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync/atomic"

	pb "github.com/converged-computing/rainbow/pkg/api/v1"
	"github.com/converged-computing/rainbow/pkg/config"
	"github.com/converged-computing/rainbow/pkg/database"
	"github.com/converged-computing/rainbow/pkg/graph/backend"
	"github.com/converged-computing/rainbow/pkg/graph/selection"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

const (
	protocol = "tcp"
	success  = "processed successfully"
)

var (
	defaultName = "rainbow"
)

// Server is used to implement your Service.
type Server struct {
	pb.UnimplementedRainbowSchedulerServer
	server   *grpc.Server
	listener net.Listener

	// counter will be for job ids
	counter     atomic.Uint64
	name        string
	version     string
	secret      string
	globalToken string
	db          *database.Database
	host        string

	// graph database handle
	graph              backend.GraphBackend
	selectionAlgorithm selection.SelectionAlgorithm
}

// NewServer creates a new "scheduler" server
// The scheduler server registers clusters and then accepts jobs
func NewServer(
	cfg *config.RainbowConfig,
	version, sqliteFile string,
	cleanup bool,
	globalToken, host string,
) (*Server, error) {

	if cfg.Scheduler.Secret == "" {
		return nil, errors.New("secret is required")
	}
	if version == "" {
		return nil, errors.New("version is required")
	}
	if cfg.Scheduler.Name == "" {
		cfg.Scheduler.Name = defaultName
	}

	// Prepare the selection algorithm
	// TODO: we probably want to allow a server to enable one or more selection
	// and match algorithms, and then the user/cluster can select from that set.
	selectAlgo, err := selection.Get(cfg.Scheduler.Algorithms.Selection.Name)
	if err != nil {
		log.Fatal(err)
	}
	err = selectAlgo.Init(cfg.Scheduler.Algorithms.Selection.Options)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("🧩️ selection algorithm: %v", selectAlgo.Name())

	// Load the graph backend!
	graphDB, err := backend.Get(cfg.GraphDatabase.Name)
	if err != nil {
		log.Fatal(err)
	}

	// Run init with any options from the config, and the match algorithm
	graphDB.Init(cfg.GraphDatabase.Options)
	log.Printf("🧩️ graph database: %v", graphDB.Name())

	// init the database, creating jobs and clusters tables
	db, err := database.InitDatabase(sqliteFile, cleanup)
	if err != nil {
		log.Fatal(err)
	}

	return &Server{
		db:                 db,
		name:               cfg.Scheduler.Name,
		graph:              graphDB,
		version:            version,
		secret:             cfg.Scheduler.Secret,
		globalToken:        globalToken,
		selectionAlgorithm: selectAlgo,
		host:               host,
	}, nil
}

func (s *Server) String() string {
	return fmt.Sprintf("%s v%s", s.name, s.version)
}

func (s *Server) GetCounter() int64 {
	return int64(s.counter.Load())
}

func (s *Server) GetName() string {
	return s.name
}

func (s *Server) GetVersion() string {
	return s.version
}

func (s *Server) Stop() {
	log.Printf("stopping server: %s", s.String())
	if s.listener != nil {
		if err := s.listener.Close(); err != nil {
			log.Printf("error closing listener: %v", err)
		}
	}
	if s.server != nil {
		s.server.Stop()
	}
}

// Start the server
func (s *Server) Start(ctx context.Context, host string) error {
	// Create a listener on the specified address.
	lis, err := net.Listen(protocol, host)
	if err != nil {
		return errors.Wrapf(err, "failed to listen: %s", host)
	}
	return s.serve(ctx, lis)
}

// serve is the main function to ensure the server is listening, etc.
// If we have an additional database to add, ensure it is added
func (s *Server) serve(_ context.Context, lis net.Listener) error {
	if lis == nil {
		return errors.New("listener is required")
	}
	s.listener = lis

	// TODO: should we add grpc.KeepaliveParams here?
	s.server = grpc.NewServer()

	// This is the main rainbow scheduler service
	pb.RegisterRainbowSchedulerServer(s.server, s)

	// Add the graph backend to it
	s.graph.RegisterService(s.server)

	log.Printf("server listening: %v", s.listener.Addr())
	if err := s.server.Serve(s.listener); err != nil && err.Error() != "closed" {
		return errors.Wrap(err, "failed to serve")
	}
	return nil
}
