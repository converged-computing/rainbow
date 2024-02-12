package client

import (
	"context"
	"time"

	pb "github.com/converged-computing/rainbow/pkg/api/v1"
	"github.com/pkg/errors"
	ts "google.golang.org/protobuf/types/known/timestamppb"
)

type RegisterRequest struct {
	Name   string
	Secret string
}

// Register makes a request to register a new cluster
func (c *RainbowClient) Register(
	ctx context.Context,
	clusterName string,
	secret string,
) (*pb.RegisterResponse, error) {

	response := &pb.RegisterResponse{}

	// TODO add secret requirement when server has database
	if clusterName == "" {
		return response, errors.New("message is required")
	}
	if !c.Connected() {
		return response, errors.New("client is not connected")
	}

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	// Hit the register endpoint
	response, err := c.service.Register(ctx, &pb.RegisterRequest{
		Name:   clusterName,
		Secret: secret,
		Sent:   ts.Now(),
	})

	// For now we blindly accept all register, it's a fake endpoint
	if err != nil {
		return response, errors.Wrap(err, "could not register cluster")
	}
	return response, nil
}
