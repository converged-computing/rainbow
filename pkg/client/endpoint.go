package client

import (
	"context"
	"fmt"
	"time"

	pb "github.com/converged-computing/rainbow/pkg/api/v1"
	"github.com/converged-computing/rainbow/pkg/types"
	"github.com/pkg/errors"
	ts "google.golang.org/protobuf/types/known/timestamppb"
)

type RegisterRequest struct {
	Name   string
	Secret string
}

// SubmitJob submits a job to a named cluster.
// The token specific to the cluster is required
func (c *RainbowClient) SubmitJob(
	ctx context.Context,
	job types.JobSpec,
	cluster string,
	token string,
) (*pb.SubmitJobResponse, error) {

	response := &pb.SubmitJobResponse{}

	// First validate the job
	if job.Nodes < 1 {
		return response, fmt.Errorf("nodes must be greater than 1")
	}
	if !c.Connected() {
		return response, errors.New("client is not connected")
	}
	if cluster == "" {
		return response, errors.New("cluster name is required")
	}

	// Contact the server...
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	// Validate that the cluster exists, and we have the right token.
	// The response is the same either way - not found does not reveal
	// additional information to the client trying to find it
	response, err := c.service.SubmitJob(ctx, &pb.SubmitJobRequest{
		Name:    job.Name,
		Token:   token,
		Nodes:   job.Nodes,
		Tasks:   job.Tasks,
		Cluster: cluster,
		Command: job.Command,
		Sent:    ts.Now(),
	})
	return response, err
}

// RequestJobs requests jobs for a specific cluster
func (c *RainbowClient) RequestJobs(
	ctx context.Context,
	cluster string,
	secret string,
	maxJobs int32,
) (*pb.RequestJobsResponse, error) {

	response := &pb.RequestJobsResponse{}
	if !c.Connected() {
		return response, errors.New("client is not connected")
	}
	if cluster == "" {
		return response, errors.New("cluster name is required")
	}
	if secret == "" {
		return response, errors.New("cluster secret is required")
	}

	// Contact the server...
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	response, err := c.service.RequestJobs(ctx, &pb.RequestJobsRequest{
		Cluster: cluster,
		Secret:  secret,
		MaxJobs: maxJobs,
		Sent:    ts.Now(),
	})
	return response, err
}

// RequestJobs requests jobs for a specific cluster
func (c *RainbowClient) AcceptJobs(
	ctx context.Context,
	cluster string,
	secret string,
	jobids []int32,
) (*pb.AcceptJobsResponse, error) {

	response := &pb.AcceptJobsResponse{}

	// First validate the job
	if len(jobids) < 1 {
		return response, fmt.Errorf("jobids to accept must be greater than 0")
	}
	if !c.Connected() {
		return response, errors.New("client is not connected")
	}
	if cluster == "" {
		return response, errors.New("cluster name is required")
	}
	if secret == "" {
		return response, errors.New("cluster secret is required")
	}

	// Contact the server...
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	response, err := c.service.AcceptJobs(ctx, &pb.AcceptJobsRequest{
		Cluster: cluster,
		Secret:  secret,
		Jobids:  jobids,
		Sent:    ts.Now(),
	})
	return response, err
}

// Register makes a request to register a new cluster
func (c *RainbowClient) Register(
	ctx context.Context,
	cluster string,
	secret string,
) (*pb.RegisterResponse, error) {

	response := &pb.RegisterResponse{}
	if cluster == "" {
		return response, errors.New("cluster is required")
	}
	if secret == "" {
		return response, errors.New("secret is required")
	}
	if !c.Connected() {
		return response, errors.New("client is not connected")
	}

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	// Hit the register endpoint
	response, err := c.service.Register(ctx, &pb.RegisterRequest{
		Name:   cluster,
		Secret: secret,
		Sent:   ts.Now(),
	})

	// For now we blindly accept all register, it's a fake endpoint
	if err != nil {
		return response, errors.Wrap(err, "could not register cluster")
	}
	return response, nil
}
