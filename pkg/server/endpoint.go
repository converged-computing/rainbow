package server

import (
	"context"
	"fmt"
	"log"

	pb "github.com/converged-computing/rainbow/pkg/api/v1"
	"github.com/converged-computing/rainbow/pkg/database"
	"github.com/converged-computing/rainbow/pkg/graph"

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

	// Cluster nodes are required
	if in.Nodes == "" {
		return nil, errors.New("cluster nodes are required")
	}

	// That can be read in...
	nodes, err := graph.ReadNodeJsonGraphString(in.Nodes)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("cluster nodes are invalid: %s", err))
	}

	log.Printf("📝️ received register: %s", in.Name)
	response, err := s.db.RegisterCluster(in.Name, s.globalToken, nodes)
	if err != nil {
		return response, err
	}

	// If we get here, now we can interact with the graph database to add the nodes
	if response.Status == pb.RegisterResponse_REGISTER_SUCCESS {
		err = s.graph.AddCluster(in.Name, &nodes, in.Subsystem)
	}
	return response, err
}

// SubmitJob submits a job to a specific cluster, or adds an entry to the database
func (s *Server) SubmitJob(_ context.Context, in *pb.SubmitJobRequest) (*pb.SubmitJobResponse, error) {
	if in == nil {
		return nil, errors.New("request is required")
	}

	// Keep a list of clusters to send to the database
	lookup := map[string]*database.Cluster{}
	clusters := []string{}

	// We submit work to one or more clusters, which must be validated via token
	// This is a very simple auth setup that needs to be improved upon, but
	// should work for a prototype
	for _, cluster := range in.Clusters {

		// No good if no name
		if cluster.Name == "" {
			log.Println("warning: cluster in request is missing a name and cannot be considered")
			continue
		}
		// No good if no token
		if cluster.Token == "" {
			log.Printf("warning: cluster %s does not have a token and cannot be considered\n", cluster.Name)
			continue
		}

		// Validate the token for the named cluster (if it exists)
		cluster, err := s.db.ValidateClusterToken(cluster.Name, cluster.Token)
		if err != nil {
			return nil, err
		}
		clusters = append(clusters, cluster.Name)
		lookup[cluster.Name] = cluster
	}

	// Only proceed if we can consider at least one cluster
	if len(clusters) == 0 {
		return nil, errors.New("one or more authenticated clusters are required")
	}

	log.Printf("📝️ received job %s for %d contender clusters", in.Name, len(clusters))

	// Use the algorithm to select a final cluster
	selected, err := s.algorithm.Select(clusters)
	if err != nil {
		return nil, err
	}
	response, err := s.db.SubmitJob(in, lookup[selected])
	if err == nil {
		log.Printf("📝️ job %s is assigned to cluster %s", in.Name, selected)
	}
	return response, err
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

// RequestJobs receives a cluster / instance / other receiving entity request for jobs
func (s *Server) AcceptJobs(_ context.Context, in *pb.AcceptJobsRequest) (*pb.AcceptJobsResponse, error) {
	if in == nil {
		return nil, errors.New("request is required")
	}

	// Nogo without a secret to validate cluster owns the namespace
	if in.Secret == "" {
		return nil, errors.New("a cluster secret is required")
	}

	// Doesn't make sense to accept < 1
	if len(in.Jobids) < 1 {
		return nil, errors.New("one or more jobs must be accepted")
	}

	// Validate the secret matches the cluster
	cluster, err := s.db.ValidateClusterSecret(in.Cluster, in.Secret)
	if err != nil {
		return nil, err
	}
	log.Printf("🌀️ accepting %d for cluster %s", len(in.Jobids), cluster.Name)
	return s.db.AcceptJobs(in, cluster)
}
