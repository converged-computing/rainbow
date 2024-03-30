package memory

// The rainbow memory backend - vanilla / prototype

import (
	"context"
	"encoding/json"
	"log"

	js "github.com/compspec/jobspec-go/pkg/jobspec/experimental"
	jgf "github.com/converged-computing/jsongraph-go/jsongraph/v2/graph"

	"github.com/converged-computing/rainbow/pkg/graph/algorithm"
	"github.com/converged-computing/rainbow/pkg/graph/backend"
	"github.com/converged-computing/rainbow/plugins/backends/memory/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// This is the global, in memory graph handle
var (
	memoryHost  = ":50051"
	graphClient *Graph
)

type MemoryGraph struct{}

var (
	description = "in-memory vanilla graph database for rainbow"
	memoryName  = "memory"
)

func (m MemoryGraph) Name() string {
	return memoryName
}

func (m MemoryGraph) Description() string {
	return description
}

// AddCluster adds a JGF graph of new nodes
// Note that a client can interact with the database (in read only)
// but since this is directly in the rainbow cluster, we call
// the functions directly. The "addCluster" here is referring
// to the dominant subsystem, while a "subsystem" below is
// considered supplementary to that.
func (m MemoryGraph) AddCluster(
	name string,
	nodes *jgf.JsonGraph,
	subsystem string,
) error {
	return graphClient.LoadClusterNodes(name, nodes, subsystem)
}

// Add subsystem adds a new subsystem to the graph!
func (m MemoryGraph) AddSubsystem(
	name string,
	nodes *jgf.JsonGraph,
	subsystem string,
) error {
	return graphClient.LoadSubsystemNodes(name, nodes, subsystem)
}

func (m MemoryGraph) RegisterService(s *grpc.Server) error {

	// This is akin to calling init
	// The service is in the same module as here, so is available to the grpc functions
	log.Printf("🧠️ Registering memory graph database...\n")

	graphClient = NewGraph()

	service.RegisterMemoryGraphServer(s, MemoryServer{})
	return nil
}

// Satisfies - determine what clusters satisfy a jobspec request
// Since this is called from the client function, it's technically
// running from the client (not from the server)
func (g MemoryGraph) Satisfies(
	jobspec *js.Jobspec,
	matcher algorithm.MatchAlgorithm,
) ([]string, error) {

	matches := []string{}
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	conn, err := grpc.Dial(memoryHost, opts...)
	if err != nil {
		return matches, err
	}
	defer conn.Close()
	client := service.NewMemoryGraphClient(conn)

	// Prepare a satisfy request, the jobspec needs to be serialized to string
	out, err := json.Marshal(jobspec)
	if err != nil {
		return matches, err
	}
	request := service.SatisfyRequest{Payload: string(out)}
	ctx := context.Background()
	response, err := client.Satisfy(ctx, &request)
	if err != nil {
		return matches, err
	}
	return response.Clusters, nil
}

// Init provides extra initialization functionality, if needed
// The in memory database can take a backup file if desired
func (g MemoryGraph) Init(
	options map[string]string,
) error {
	backupFile, ok := options["backupFile"]
	if ok {
		graphClient.backupFile = backupFile
	}

	quiet, ok := options["quiet"]
	if ok {
		if quiet == "true" || quiet == "yes" {
			graphClient.quiet = true
		}
	}

	// Warning: this assumes one client running with one graph host
	host, ok := options["host"]
	if ok {
		memoryHost = host
	}
	return nil
}

// Add the backend to be known to rainbow
func init() {

	graph := MemoryGraph{}
	backend.Register(graph)
}
