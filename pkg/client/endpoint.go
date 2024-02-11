package client

import (
	"context"
	"fmt"
	"time"

	pb "github.com/converged-computing/rainbow/pkg/api/v1"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/anypb"
	ts "google.golang.org/protobuf/types/known/timestamppb"
	wrapper "google.golang.org/protobuf/types/known/wrapperspb"
)

// Register makes a request to register a new cluster
func (c *SimpleClient) Register(
	ctx context.Context,
	clusterName string,
	secret string,
) (string, error) {

	// TODO add secret requirement when server has database
	if clusterName == "" {
		return "", errors.New("message is required")
	}
	if !c.Connected() {
		return "", errors.New("client is not connected")
	}

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	// create cluster name with wrapper
	a, err := anypb.New(wrapper.String(clusterName))
	if err != nil {
		return "", errors.Wrap(err, "unable to create message")
	}

	// Hit the register endpoint
	r, err := c.service.Register(ctx, &pb.Request{
		Content: &pb.Content{
			Id:   uuid.New().String(),
			Data: a,
		},
		Sent: ts.Now(),
	})

	// For now we blindly accept all register, it's a fake endpoint
	if err != nil {
		return "register failed", errors.Wrap(err, "could not register cluster")
	}
	fmt.Println(r)
	return "register success", nil
}
