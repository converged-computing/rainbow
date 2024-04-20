package client

import (
	"context"
	"log"
	"time"

	js "github.com/compspec/jobspec-go/pkg/nextgen/v1"

	pb "github.com/converged-computing/rainbow/pkg/api/v1"
	"github.com/converged-computing/rainbow/pkg/certs"
	"github.com/converged-computing/rainbow/pkg/config"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// RainbowClient is our instantiation of Client
type RainbowClient struct {
	host       string
	connection *grpc.ClientConn
	service    pb.RainbowSchedulerClient
}

var _ Client = (*RainbowClient)(nil)
var defaultTimeout = 200 * time.Second

// Client interface defines functions required for a valid client
type Client interface {

	// Cluster interactions
	Register(ctx context.Context, clusterName, secret, clusterNodes, subsystem string) (*pb.RegisterResponse, error)
	RegisterSubsystem(ctx context.Context, clusterName, secret, subsystemNodes, subsystem string) (*pb.RegisterResponse, error)

	// Update
	UpdateState(ctx context.Context, clusterName, secret, stateFile string) (*pb.UpdateStateResponse, error)

	// Job Client Interactions
	AcceptJobs(ctx context.Context, cluster, secret string, jobids []int32) (*pb.AcceptJobsResponse, error)
	SubmitJob(ctx context.Context, job *js.Jobspec, cfg *config.RainbowConfig) (*pb.SubmitJobResponse, error)
	ReceiveJobs(ctx context.Context, cluster, token string, maxJobs int32) (*pb.ReceiveJobsResponse, error)
}

// NewClient creates a new RainbowClient
func NewClient(host string, cert *certs.Certificate) (Client, error) {
	if host == "" {
		return nil, errors.New("host is required")
	}

	log.Printf("🌈️ starting client (%s)...", host)
	c := &RainbowClient{host: host}

	// The context allows us to control the timeout
	// ctx, cancel := context.WithTimeout(context.TODO(), defaultTimeout)
	// defer cancel()

	// Set up a connection to the server.
	// creds := grpc.WithTransportCredentials(insecure.NewCredentials())
	// conn, err := grpc.DialContext(ctx, c.GetHost(), creds, grpc.WithBlock())
	// Are we using tls?
	var transportCreds credentials.TransportCredentials
	var err error
	if !cert.IsEmpty() {
		log.Printf("🔐️ adding tls credentials")
		transportCreds = cert.GetClientCredentials()
	} else {
		transportCreds = insecure.NewCredentials()
	}

	// Set up a connection to the server.
	creds := grpc.WithTransportCredentials(transportCreds)
	conn, err := grpc.Dial(c.GetHost(), creds, grpc.WithBlock())	

	// Something else I was trying
	// conn, err := grpc.Dial(c.GetHost(),
	//	creds,
		//		grpc.WithBlock(),
	//	grpc.FailOnNonTempDialError(true),
	// )
	if err != nil {
		return nil, errors.Wrapf(err, "unable to connect to %s", host)
	}

	c.connection = conn
	c.service = pb.NewRainbowSchedulerClient(conn)

	return c, nil
}

// Close closes the created resources (e.g. connection).
func (c *RainbowClient) Close() error {
	if c.connection != nil {
		return c.connection.Close()
	}
	return nil
}

// Connected returns  true if we are connected and the connection is ready
func (c *RainbowClient) Connected() bool {
	return c.service != nil && c.connection != nil && c.connection.GetState() == connectivity.Ready
}

// GetHost returns the private hostn name
func (c *RainbowClient) GetHost() string {
	return c.host
}
