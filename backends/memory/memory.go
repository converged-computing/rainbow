package memory

// The rainbow memory backend - vanilla / prototype

import (
	"log"

	jgf "github.com/converged-computing/jsongraph-go/jsongraph/v2/graph"

	"github.com/converged-computing/rainbow/backends/memory/service"
	"github.com/converged-computing/rainbow/pkg/graph/backend"
	"google.golang.org/grpc"
)

// This is the global, in memory graph handle
var graphClient *ClusterGraph

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
// the functions directly.
func (m MemoryGraph) AddCluster(name string, nodes *jgf.JsonGraph) error {

	// How this might look for an external client
	/* var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	conn, err := grpc.Dial("127.0.0.1:50051", opts...)
	if err != nil {
		return err
	}
	defer conn.Close()
	client := service.NewMemoryGraphClient(conn)
	ctx := context.Background()
	client.Register(...) */
	return graphClient.LoadClusterNodes(name, nodes)
}

func (m MemoryGraph) RegisterService(s *grpc.Server) error {

	// This is akin to calling init
	// The service is in the same module as here, so is available to the grpc functions
	log.Printf("🧠️ Registering memory graph database...\n")
	graphClient = NewClusterGraph()

	service.RegisterMemoryGraphServer(s, MemoryServer{})
	return nil
}

// Add the backend to be known to libpak
func init() {

	graph := MemoryGraph{}
	backend.Register(graph)
}

// Satisfies - determine what clusters satisfy a jobspec request
func (g MemoryGraph) Satisfies(jobspec string) error {
	return nil
}

// Init provides extra initialization functionality, if needed
// The in memory database can take a backup file if desired
func (g MemoryGraph) Init(options map[string]string) error {
	backupFile, ok := options["backupFile"]
	if ok {
		graphClient.backupFile = backupFile
	}
	return nil
}
