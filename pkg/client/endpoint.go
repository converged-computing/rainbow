package client

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	js "github.com/compspec/jobspec-go/pkg/nextgen/v1"
	pb "github.com/converged-computing/rainbow/pkg/api/v1"
	"github.com/converged-computing/rainbow/pkg/config"
	"github.com/converged-computing/rainbow/pkg/graph"
	"github.com/converged-computing/rainbow/pkg/graph/algorithm"
	"github.com/converged-computing/rainbow/pkg/graph/backend"
	"github.com/converged-computing/rainbow/pkg/utils"
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
	job *js.Jobspec,
	cfg *config.RainbowConfig,
) (*pb.SubmitJobResponse, error) {

	response := &pb.SubmitJobResponse{}
	if !c.Connected() {
		return response, errors.New("client is not connected")
	}
	if len(cfg.Clusters) == 0 {
		return response, errors.New("one or more clusters must be defined in the configuration file")
	}

	// Request work directly to the database
	graphDB, err := backend.Get(cfg.GraphDatabase.Name)
	if err != nil {
		return response, err
	}

	// Prepare the subsystem match algorithm
	matchAlgo, err := algorithm.Get(cfg.Scheduler.Algorithms.Match.Name)
	if err != nil {
		log.Fatal(err)
	}
	matchAlgo.Init(cfg.Scheduler.Algorithms.Match.Options)

	// TODO we need to have a check here to see what clusters
	// the user has permission to do. Either that can be represented in
	// the graph database (and the call goes directly to it) or it
	// is checked first in rainbow, and still enforced in the graph
	// (but we limit our search). Likely the first is preferable.
	// Ask the graphDB if the jobspec can be satisfied
	// TODO what does a match look like?
	matches, err := graphDB.Satisfies(job, matchAlgo)
	if err != nil {
		return response, err
	}

	// Cut out early (without contacting rainbow) if there are no matches
	if len(matches) > 0 {
		log.Printf("🎯️ We found %d matches! %s\b", len(matches), matches)
	} else {
		return response, fmt.Errorf("😥️ There were no matches for this job")
	}
	// Now contact the rainbow server with clusters...
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	// Prepare clusters for submit jobs request
	clusters := make([]*pb.SubmitJobRequest_Cluster, len(cfg.Clusters))

	// Take an intersection of clusters and matches
	// A token will not be returned if we do not know about the cluster
	for i, match := range matches {
		creds := cfg.GetClusterToken(match)
		if creds != "" {
			clusters[i] = &pb.SubmitJobRequest_Cluster{Token: creds, Name: match}
		}
	}

	// Jobspec gets converted back to string for easier serialization
	out, err := job.JobspecToYaml()
	if err != nil {
		return response, err
	}

	// Validate that the cluster exists, and we have the right token.
	// The response is the same either way - not found does not reveal
	// additional information to the client trying to find it
	response, err = c.service.SubmitJob(ctx, &pb.SubmitJobRequest{
		Name:     job.GetJobName(),
		Clusters: clusters,
		Jobspec:  string(out),
		Sent:     ts.Now(),
	})
	return response, err
}

// ReceiveJobs (request them) for a specific clusters
func (c *RainbowClient) ReceiveJobs(
	ctx context.Context,
	cluster string,
	secret string,
	maxJobs int32,
) (*pb.ReceiveJobsResponse, error) {
	response := &pb.ReceiveJobsResponse{}
	if !c.Connected() {
		return response, errors.New("client is not connected")
	}
	if cluster == "" {
		return response, errors.New("cluster name is required")
	}
	if secret == "" {
		return response, errors.New("cluster secret is required")
	}

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	response, err := c.service.ReceiveJobs(ctx, &pb.ReceiveJobsRequest{
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
	clusterNodes string,
	subsystem string,
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
	// The cluster nodes file must be defined
	if clusterNodes == "" {
		return response, fmt.Errorf("cluster nodes file must be provided with --nodes-json")
	}

	// and exist
	_, err := utils.PathExists(clusterNodes)
	if err != nil {
		return response, errors.New(fmt.Sprintf("cluster nodes file %s does not exist: %s", clusterNodes, err))
	}

	// Read in the cluster nodes
	// jgf: is the struct, we do this to ensure it can unmarshal
	// nodes: is the string to send over gRPC
	_, nodes, err := graph.ReadNodeJsonGraph(clusterNodes)
	if err != nil {
		return response, err
	}

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	// Hit the register endpoint
	response, err = c.service.Register(ctx, &pb.RegisterRequest{
		Name:      cluster,
		Secret:    secret,
		Nodes:     nodes,
		Subsystem: subsystem,
		Sent:      ts.Now(),
	})

	// For now we blindly accept all register, it's a fake endpoint
	if err != nil {
		return response, errors.Wrap(err, "could not register cluster")
	}
	return response, nil
}

// UpdateState of an existing cluster
func (c *RainbowClient) UpdateState(
	ctx context.Context,
	cluster string,
	secret string,
	stateFile string,
) (*pb.UpdateStateResponse, error) {

	response := &pb.UpdateStateResponse{}

	// Unlike register, this is the cluster to add the subsytem for.
	if cluster == "" {
		return response, errors.New("cluster is required")
	}
	if secret == "" {
		return response, errors.New("secret is required")
	}
	if !c.Connected() {
		return response, errors.New("client is not connected")
	}
	if stateFile == "" {
		return response, fmt.Errorf("a state file must be provided with --state-file")
	}
	_, err := utils.PathExists(stateFile)
	if err != nil {
		return response, errors.New(fmt.Sprintf("state file %s does not exist: %s", stateFile, err))
	}

	// Read stateFile into payload
	states, err := os.ReadFile(stateFile)
	if err != nil {
		return response, err
	}

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	// Hit the register subsystem endpoint
	response, err = c.service.UpdateState(ctx, &pb.UpdateStateRequest{
		Cluster: cluster,
		Secret:  secret,
		Payload: string(states),
	})
	if err != nil {
		return response, errors.Wrap(err, "could not update endpoint")
	}
	return response, nil
}

// Register makes a request to register a new cluster
func (c *RainbowClient) RegisterSubsystem(
	ctx context.Context,
	cluster string,
	secret string,
	subsystemNodes string,
	subsystem string,
) (*pb.RegisterResponse, error) {

	response := &pb.RegisterResponse{}

	// Unlike register, this is the cluster to add the subsytem for.
	if cluster == "" {
		return response, errors.New("cluster is required")
	}
	if secret == "" {
		return response, errors.New("secret is required")
	}
	if !c.Connected() {
		return response, errors.New("client is not connected")
	}
	if subsystemNodes == "" {
		return response, fmt.Errorf("subsystem nodes file must be provided with --subsys-nodes")
	}
	_, err := utils.PathExists(subsystemNodes)
	if err != nil {
		return response, errors.New(fmt.Sprintf("subsystem nodes file %s does not exist: %s", subsystemNodes, err))
	}

	// Read in the subsystem nodes - still JGF!
	_, nodes, err := graph.ReadNodeJsonGraph(subsystemNodes)
	if err != nil {
		return response, err
	}

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	// Hit the register subsystem endpoint
	response, err = c.service.RegisterSubsystem(ctx, &pb.RegisterRequest{
		Name:      cluster,
		Secret:    secret,
		Nodes:     nodes,
		Subsystem: subsystem,
		Sent:      ts.Now(),
	})

	// For now we blindly accept all register, it's a fake endpoint

	if err != nil {
		return response, errors.Wrap(err, "could not register cluster")
	}

	return response, nil
}
