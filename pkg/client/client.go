package client

import (
	"context"
	"log"

	pb "github.com/converged-computing/rainbow/pkg/api/v1"
	"github.com/converged-computing/rainbow/pkg/provider"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
)

var _ Client = (*SimpleClient)(nil)

type Client interface {
	Serial(ctx context.Context, message string) (string, error)
	Stream(ctx context.Context, it provider.MessageIterator) error
	Register(ctx context.Context, clusterName, secret string) (string, error)
}

func NewClient(target string) (Client, error) {
	if target == "" {
		return nil, errors.New("target is required")
	}

	log.Printf("starting client (%s)...", target)

	c := &SimpleClient{
		target: target,
	}

	// Set up a connection to the server.
	creds := grpc.WithTransportCredentials(insecure.NewCredentials())
	conn, err := grpc.Dial(c.GetTarget(), creds, grpc.WithBlock())
	if err != nil {
		return nil, errors.Wrapf(err, "unable to connect to %s", target)
	}

	c.connection = conn
	c.service = pb.NewServiceClient(conn)

	return c, nil
}

type SimpleClient struct {
	target     string
	connection *grpc.ClientConn
	service    pb.ServiceClient
}

// Close closes the created resources (e.g. connection).
func (c *SimpleClient) Close() error {
	if c.connection != nil {
		return c.connection.Close()
	}
	return nil
}

func (c *SimpleClient) Connected() bool {
	return c.service != nil && c.connection != nil && c.connection.GetState() == connectivity.Ready
}

func (c *SimpleClient) GetTarget() string {
	return c.target
}
