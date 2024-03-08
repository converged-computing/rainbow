package database

import (
	"fmt"

	"github.com/converged-computing/jsongraph-go/jsongraph/v2/graph"
	pb "github.com/converged-computing/rainbow/pkg/api/v1"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"

	"log"
)

// RegisterCluster registers a cluster or returns another status
func (db *Database) RegisterCluster(
	name, globalToken string,
	nodesGraph graph.JsonGraph,
) (*pb.RegisterResponse, error) {

	response := &pb.RegisterResponse{}

	// Verify we have the graph
	log.Printf("Received cluster graph with %d nodes and %d edges\n", len(nodesGraph.Graph.Nodes), len(nodesGraph.Graph.Edges))

	// Connect!
	conn, err := db.connect()
	if err != nil {
		return response, err
	}
	defer conn.Close()

	// First determine if it exists - this needs to get the results
	query := fmt.Sprintf("SELECT count(*) from clusters WHERE name = '%s'", name)
	count, err := countResults(conn, query)
	if err != nil {
		return response, err
	}
	// Debugging extra for now
	log.Printf("%s: (%d)\n", query, count)

	// Case 1: already exists
	if count > 0 {
		response.Status = pb.RegisterResponse_REGISTER_EXISTS
		return response, nil
	}

	// Generate a token and a secret.
	// The "token" is given to clients to submit jobs
	// If we are using a global token, they are given the same one
	var token string
	if globalToken != "" {
		token = globalToken
	} else {
		token = uuid.New().String()
	}

	// The "secret" is used by the cluster to request jobs for itself
	secret := uuid.New().String()
	values := fmt.Sprintf("(\"%s\", \"%s\", \"%s\")", name, token, secret)
	query = fmt.Sprintf("INSERT into clusters (name, token, secret) VALUES %s", values)
	result, err := conn.Exec(query)

	// Error with request
	if err != nil {
		response.Status = pb.RegisterResponse_REGISTER_ERROR
		return response, err
	}
	count, err = result.RowsAffected()
	log.Printf("%s: (%d)\n", query, count)

	// REGISTER_SUCCESS - the only case to pass forward credentials
	if count > 0 {
		response.Status = pb.RegisterResponse_REGISTER_SUCCESS
		response.Token = token
		response.Secret = secret
		return response, nil
	}

	// REGISTER_ERROR
	response.Status = pb.RegisterResponse_REGISTER_ERROR
	return response, err
}
