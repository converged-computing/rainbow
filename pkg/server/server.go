package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync/atomic"

	pb "github.com/converged-computing/rainbow/pkg/api/v1"
	"github.com/converged-computing/rainbow/pkg/database"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

const (
	protocol = "tcp"
	success  = "processed successfully"
)

var (
	defaultEnv  = "development"
	defaultName = "rainbow"
)

// NewServer creates a new "scheduler" server
// The scheduler server registers clusters and then accepts jobs
func NewServer(
	name, version, environment, sqliteFile string,
	cleanup bool,
	secret string,
) (*Server, error) {

	if secret == "" {
		return nil, errors.New("secret is required")
	}
	if version == "" {
		return nil, errors.New("version is required")
	}
	if name == "" {
		name = defaultName
	}
	if environment == "" {
		environment = defaultEnv
	}

	// init the database, creating jobs and clusters tables
	db, err := database.InitDatabase(sqliteFile, cleanup)
	if err != nil {
		log.Fatal(err)
	}

	return &Server{
		db:          db,
		name:        name,
		version:     version,
		secret:      secret,
		environment: environment,
	}, nil
}

// Server is used to implement your Service.
type Server struct {
	pb.UnimplementedRainbowSchedulerServer
	server   *grpc.Server
	listener net.Listener

	// counter will be for job ids
	counter     atomic.Uint64
	name        string
	version     string
	environment string
	secret      string
	db          *database.Database
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
func (s *Server) Start(ctx context.Context, address string) error {
	// Create a listener on the specified address.
	lis, err := net.Listen(protocol, address)
	if err != nil {
		return errors.Wrapf(err, "failed to listen: %s", address)
	}
	return s.serve(ctx, lis)
}

// serve is the main function to ensure the server is listening, etc.
func (s *Server) serve(_ context.Context, lis net.Listener) error {
	if lis == nil {
		return errors.New("listener is required")
	}
	s.listener = lis
	s.server = grpc.NewServer()
	pb.RegisterRainbowSchedulerServer(s.server, s)
	log.Printf("server listening: %v", s.listener.Addr())
	if err := s.server.Serve(s.listener); err != nil && err.Error() != "closed" {
		return errors.Wrap(err, "failed to serve")
	}
	return nil
}
