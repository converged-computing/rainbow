package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync/atomic"

	pb "github.com/converged-computing/rainbow/pkg/api/v1"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

const (
	protocol = "tcp"
	success  = "processed successfully"
)

// NewServer creates a new "scheduler" server
// The scheduler server registers clusters and then accepts jobs
func NewServer(name, version, environment string) (*Server, error) {
	if name == "" {
		return nil, errors.New("name is required")
	}
	if version == "" {
		return nil, errors.New("version is required")
	}
	if environment == "" {
		return nil, errors.New("environment is required")
	}

	return &Server{
		name:        name,
		version:     version,
		environment: environment,
	}, nil
}

// Server is used to implement your Service.
type Server struct {
	pb.UnimplementedServiceServer
	server      *grpc.Server
	listener    net.Listener
	counter     atomic.Uint64 // counter for messages
	name        string        // server name
	version     string        // server version
	environment string        // server environment
}

func (s *Server) String() string {
	return fmt.Sprintf("%s (%s) v%s", s.name, s.environment, s.version)
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

func (s *Server) GetEnvironment() string {
	return s.environment
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

func (s *Server) Start(ctx context.Context, address string) error {
	// Create a listener on the specified address.
	lis, err := net.Listen(protocol, address)
	if err != nil {
		return errors.Wrapf(err, "failed to listen: %s", address)
	}
	return s.serve(ctx, lis)
}

func (s *Server) serve(_ context.Context, lis net.Listener) error {
	if lis == nil {
		return errors.New("listener is required")
	}
	s.listener = lis
	s.server = grpc.NewServer()
	pb.RegisterServiceServer(s.server, s)
	log.Printf("server listening: %v", s.listener.Addr())
	if err := s.server.Serve(s.listener); err != nil && err.Error() != "closed" {
		return errors.Wrap(err, "failed to serve")
	}
	return nil
}
