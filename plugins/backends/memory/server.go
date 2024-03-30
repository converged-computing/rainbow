package memory

import (
	"context"

	"github.com/converged-computing/rainbow/pkg/config"
	"github.com/converged-computing/rainbow/pkg/graph/algorithm"
	"github.com/converged-computing/rainbow/plugins/backends/memory/service"
)

type MemoryServer struct {
	service.UnimplementedMemoryGraphServer
}

// Register takes a cluster node payload and adds to the in memory graph
func (MemoryServer) Register(c context.Context, req *service.RegisterRequest) (*service.Response, error) {
	response, err := graphClient.RegisterCluster(req.Name, req.Payload, req.Subsystem)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// Satisfy determines if the graph can satisfy a request
func (MemoryServer) Satisfy(c context.Context, req *service.SatisfyRequest) (*service.SatisfyResponse, error) {
	if req.Matcher == "" {
		req.Matcher = config.DefaultMatchAlgorithm
	}
	// Instantiate the matcher
	matcher, err := algorithm.Get(req.Matcher)
	if err != nil {
		return nil, err
	}
	response, err := graphClient.Satisfies(req.Payload, matcher)
	if err != nil {
		return nil, err
	}
	return response, nil
}
