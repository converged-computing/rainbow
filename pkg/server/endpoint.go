package server

import (
	"context"
	"log"

	pb "github.com/converged-computing/rainbow/pkg/api/v1"

	"github.com/pkg/errors"
)

// Register a new cluster with the server
func (s *Server) Register(_ context.Context, in *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	if in == nil {
		return nil, errors.New("request is required")
	}

	// Validate the secret
	if in.Secret == "" || (in.Secret != s.secret) {
		return nil, errors.New("request denied")
	}
	log.Printf("📝️ received register: %s", in.Name)
	return s.db.RegisterCluster(in.Name)
}

// SubmitJob submits a job to a specific cluster, or adds an entry to the database
func (s *Server) SubmitJob(_ context.Context, in *pb.SubmitJobRequest) (*pb.SubmitJobResponse, error) {
	if in == nil {
		return nil, errors.New("request is required")
	}

	// Nogo without a token
	if in.Token == "" {
		return nil, errors.New("a cluster token is required")
	}

	// Validate the token for the cluster (if it exists)
	cluster, err := s.db.ValidateClusterToken(in.Cluster, in.Token)
	if err != nil {
		return nil, err
	}
	log.Printf("📝️ received job %s for cluster %s", in.Name, cluster.Name)
	return s.db.SubmitJob(in, cluster)
}

// RequestJobs receives a cluster / instance / other receiving entity request for jobs
func (s *Server) RequestJobs(_ context.Context, in *pb.RequestJobsRequest) (*pb.RequestJobsResponse, error) {
	if in == nil {
		return nil, errors.New("request is required")
	}

	// Nogo without a secret to validate cluster owns the namespace
	if in.Secret == "" {
		return nil, errors.New("a cluster secret is required")
	}

	// Validate the secret matches the cluster
	cluster, err := s.db.ValidateClusterSecret(in.Cluster, in.Secret)
	if err != nil {
		return nil, err
	}
	log.Printf("🌀️ requesting %d max jobs for cluster %s", in.MaxJobs, cluster.Name)
	return s.db.RequestJobs(in, cluster)
}

// TEST ENDPOINTS ------------------------------------------------
// Stream implements the Stream method of the Service.
func (s *Server) Stream(stream pb.RainbowScheduler_StreamServer) error {
	if stream == nil {
		return errors.New("stream is required")
	}

	for {
		in, err := stream.Recv()
		if err != nil {
			return errors.Wrap(err, "failed to receive")
		}

		c := in.GetContent()
		log.Printf("received stream: %v", c.GetData())

		s.counter.Add(1)
		response := &pb.Response{
			RequestId:         c.GetId(),
			MessageCount:      s.GetCounter(),
			MessagesProcessed: s.GetCounter(),
			ProcessingDetails: success,
		}
		if err := stream.Send(response); err != nil {
			return errors.Wrap(err, "failed to send")
		}
	}
}

// Serialimplements the single method of the Service.
func (s *Server) Serial(_ context.Context, in *pb.Request) (*pb.Response, error) {
	if in == nil {
		return nil, errors.New("request is required")
	}

	c := in.GetContent()
	log.Printf("received serial: %v", c.GetData())

	// This is redundant, but for protobuf responses I like to see them directly
	return &pb.Response{
		RequestId:         c.GetId(),
		MessageCount:      s.GetCounter(),
		MessagesProcessed: s.GetCounter(),
		ProcessingDetails: success,
	}, nil
}
