package memgraph

// The rainbow memory backend - vanilla / prototype

import (
	"encoding/json"
	"log"
	"strings"

	"context"
	"fmt"

	"github.com/converged-computing/rainbow/pkg/graph"
	"github.com/converged-computing/rainbow/pkg/graph/algorithm"
	"github.com/converged-computing/rainbow/pkg/graph/backend"
	rlog "github.com/converged-computing/rainbow/pkg/logger"
	"github.com/converged-computing/rainbow/pkg/types"
	"github.com/converged-computing/rainbow/pkg/utils"
	"github.com/converged-computing/rainbow/plugins/algorithms/shared"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"google.golang.org/grpc"

	js "github.com/compspec/jobspec-go/pkg/nextgen/v1"
	v1 "github.com/compspec/jobspec-go/pkg/nextgen/v1"
	jgf "github.com/converged-computing/jsongraph-go/jsongraph/v2/graph"
)

type Memgraph struct{}

var (
	description  = "memgraph backend"
	memoryName   = "memgraph"
	memoryHost   = "bolt://localhost:7687"
	databaseName = ""
	username     = "rainbow"
	password     = "chocolate-cookies"
)

func (m Memgraph) Name() string {
	return memoryName
}

func (m Memgraph) Description() string {
	return description
}

// AddCluster adds a new cluster to the graph
// Name is the name of the cluster
func (m Memgraph) AddCluster(
	name string,
	nodes *jgf.JsonGraph,
	subsystem string,
) error {
	// Add a cluster subsystem
	if subsystem == "" {
		subsystem = types.DefaultDominantSubsystem
	}
	return m.AddSubsystem(name, nodes, subsystem)
}

// UpdateState updates the state of a cluster in memgraph
func (m Memgraph) UpdateState(
	name string,
	payload string,
) error {
	// Load state into interface
	state := types.ClusterState{}
	err := json.Unmarshal([]byte(payload), &state)
	if err != nil {
		return err
	}
	return nil
}

// GetStates for a list of clusters
func (m Memgraph) GetStates(names []string) (map[string]types.ClusterState, error) {
	return map[string]types.ClusterState{}, nil
}

// AddSusbsystem to the graph
// Name is the name of the cluster that the subsystem belongs to
func (m Memgraph) AddSubsystem(
	name string,
	nodes *jgf.JsonGraph,
	subsystem string,
) error {

	// Connect to the driver
	driver, err := neo4j.NewDriverWithContext(memoryHost, neo4j.BasicAuth(username, password, databaseName))

	if err != nil {
		return err
	}
	ctx := context.Background()
	defer driver.Close(ctx)
	err = driver.VerifyConnectivity(ctx)

	if err != nil {
		return nil
	}

	// We will need the dominant (containment) subsystem name for external edges
	// e.g., cluster-keebler-<some-id>
	domName := graph.GetNamespacedName("cluster", name)

	// Names are always prefixed with subsystem, e.g,
	// cluster-keebler
	// io-keebler
	name = fmt.Sprintf("%s-%s", subsystem, name)

	// Check that we don't have it already - a subsystem (or cluster) can only be added once
	// type likely isn't needed, but it would allow us to filter down quickly to an entire kind
	// of subsystem if needed
	query := fmt.Sprintf("MATCH (n:Subsystem{name: '%s'}) RETURN n;", name)
	rlog.Debug(query)
	result, err := neo4j.ExecuteQuery(
		ctx, driver, query, nil, neo4j.EagerResultTransformer,
		neo4j.ExecuteQueryWithDatabase(databaseName),
	)

	if err != nil {
		return err
	}
	if len(result.Records) > 0 {
		return fmt.Errorf("subsystem '%s' with type '%s' already exists", name, subsystem)
	}

	// Create the subsystem and relationships of nodes to it
	subsystem_nodes := []string{
		fmt.Sprintf("CREATE (n:Subsystem {type: '%s', name:'%s'});", subsystem, name),
	}
	relationships := []string{}
	lookup := map[string]string{}

	// Create a session
	session := driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: databaseName})
	defer session.Close(ctx)

	// Let's be pedantic - no clusters allowed without nodes or edges
	_, _, err = graph.ValidateNodes(nodes)
	if err != nil {
		return err
	}

	// Now loop through the nodes and add them, keeping a temporary lookup
	for nid, node := range nodes.Graph.Nodes {

		// Currently we are saving the type, size, and unit
		resource := types.NewResource(node)

		// Defining a lookup name means that we keep a direct index to the node in
		// the subsystem lookup. We do this for edges between subsystems so
		// they are always namespaced
		// - name with subsystem, cluster name, and original id is indexed
		//   e.g., cluster-keebler-0
		lookupName := graph.GetNamespacedName(name, nid)

		// For now I'm putting the subsystem as an attribute instead of a relation, this could change
		newNode := fmt.Sprintf(
			"CREATE (n:Node {name:'%s', type:'%s', size: '%d', unit: '%s', subsystem: '%s'});",
			lookupName,
			resource.Type,
			resource.Size,
			resource.Unit,
			subsystem,
		)
		subsystem_nodes = append(subsystem_nodes, newNode)

		// This stores the original JGF id so we can reference it for internal edge
		lookup[nid] = lookupName
	}

	// Create nodes
	rlog.Debugf("Creating %d nodes for subsystem\n", len(subsystem_nodes), subsystem)
	for _, node := range subsystem_nodes {
		_, err = neo4j.ExecuteQuery(ctx, driver, node, nil, neo4j.EagerResultTransformer, neo4j.ExecuteQueryWithDatabase(databaseName))
		if err != nil {
			return err
		}
	}

	// Create relationships
	for _, rel := range relationships {
		_, err = neo4j.ExecuteQuery(ctx, driver, rel, nil, neo4j.EagerResultTransformer, neo4j.ExecuteQueryWithDatabase(databaseName))
		if err != nil {
			return err
		}
	}

	// Now add edges
	// Count dominant vertices references
	count := 0

	// reset relationships to use for adding edges
	relationships = []string{}

	// Now add edges
	for _, edge := range nodes.Graph.Edges {

		// We are currently just saving one direction "x contains y"
		if edge.Relation != types.ContainsRelation {
			continue
		}

		// Two cases:
		// 1. the src is in the dominant subsystem
		// 2. The src is not, and both node are defined in the graph here
		subIdx1, ok1 := lookup[edge.Source]
		subIdx2, ok2 := lookup[edge.Target]
		fmt.Println(lookup)

		// Case 1: both are in the subsystem graph
		if ok1 && ok2 {
			// This says "subsystem resource in node"
			fmt.Printf("Adding internal edge for %s to %s\n", subIdx1, subIdx2)

			// Tie the node to the subsystem
			relation := fmt.Sprintf(
				"MATCH (a:Node {name: '%s'}),(b:Node {name: '%s'}) CREATE (a)-[r:%s]->(b);",
				subIdx1,
				subIdx2,
				types.ContainsRelation,
			)
			rlog.Debug(relation)
			relationships = append(relationships, relation)

		} else if ok2 {

			// Case 2: the src is in the dominant subsystem
			// We need the namespaced name for the dom lookup
			lookupName := graph.GetNamespacedName(domName, edge.Source)
			fmt.Printf("Adding dominant subsystem edge for %s to %s in %s\n", lookupName, subIdx2, subsystem)
			count += 1
			// Now add the link... the node exists in the subsystem but references a
			// different subsystem as the edge.
			// This says "dominant subsystem node conatains subsystem resource"
			relation := fmt.Sprintf(
				"MATCH (a:Node {name: '%s'}),(b:Node {name: '%s'}) CREATE (a)-[r:%s]->(b);",
				lookupName,
				subIdx2,
				types.ContainsRelation,
			)
			rlog.Debug(relation)
			relationships = append(relationships, relation)
		} else {
			return fmt.Errorf("edge %s->%s is not internal, and not connected to the dominant subsystem", edge.Source, edge.Target)
		}
	}

	// Create edges between nodes
	for _, rel := range relationships {
		_, err = neo4j.ExecuteQuery(ctx, driver, rel, nil, neo4j.EagerResultTransformer, neo4j.ExecuteQueryWithDatabase(databaseName))
		if err != nil {
			return err
		}
	}

	if count > 0 {
		log.Printf("We have made a memgraph (subsystem %s) with %d vertices, with %d connections to the dominant!", subsystem, len(subsystem_nodes), count)
	} else {
		log.Printf("We have made a memgraph (subsystem %s) with %d vertices", subsystem, len(subsystem_nodes))
	}
	return nil
}

// RegisterService does a test connection
func (m Memgraph) RegisterService(s *grpc.Server) error {

	// This is akin to calling init
	log.Printf("🧠️ Registering memgraph graph database...\n")

	// Do a test connection
	driver, err := neo4j.NewDriverWithContext(memoryHost, neo4j.BasicAuth(username, password, databaseName))
	if err != nil {
		return err
	}

	ctx := context.Background()
	defer driver.Close(ctx)
	err = driver.VerifyConnectivity(ctx)
	if err != nil {
		return err
	}

	// Prepare to have subsystems (a cluster is also a subsystem)
	indexes := []string{
		"CREATE INDEX ON :Subsystem(name);",
		"CREATE INDEX ON :Node(name);",
	}

	// Create indices
	session := driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: databaseName})
	defer session.Close(ctx)

	// Run index queries via implicit auto-commit transaction
	for _, index := range indexes {
		_, err = session.Run(ctx, index, nil)
		if err != nil {
			return err
		}
	}
	return nil
}

// Satisfies - determine what clusters satisfy a jobspec request
// Since this is called from the client function, it's technically
// running from the client (not from the server). See examples
// in comments below.
func (g Memgraph) Satisfies(
	jobspec *js.Jobspec,
	matcher algorithm.MatchAlgorithm,
) ([]string, error) {

	// Note that this algorithm is different in that it skips
	// hieuristics (checking totals) because we minimize queries
	// to the graph.
	matches := []string{}

	// Get resources that need scheduling from the jobspec
	// This is a map[string]Resource{} that may or may not have type slot
	resources := jobspec.GetScheduledNamedSlots()

	// Show the JobSpec for debugging
	fmt.Println(jobspec.JobspecToYaml())

	// Each schedulable unit will get a separate query
	var query string

	// Parse into resource structure and update the query appropriately
	var updateQuery func(resource v1.Resource, resourceTypes []string, lastSeen string) error
	updateQuery = func(resource v1.Resource, resourceTypes []string, lastSeen string) error {

		// Keep track of seen types. When we parse a slot, we need to start
		// at where we left off
		seenTypes := []string{}

		// Get local resource needs (subsystem edges), add to query
		resourceNeeds := shared.GetResourceNeeds(&resource)

		for _, resourceType := range resourceTypes {
			subsystemNeeds, exists := resourceNeeds.Subsystems[resourceType]

			// Recursive call needs to start where we left off
			seenTypes = append(seenTypes, resourceType)

			// We need to keep track of the containment edge name
			// This is how we count slots at the end
			resourceEdgeName := fmt.Sprintf("%sEdge", resourceType)

			// Here is an example asking for a subsystem attribute
			// MATCH (cluster:Node {subsystem: 'cluster', type: 'cluster'})
			//	-[r0:contains]-(rack:Node {subsystem: 'cluster', type: 'rack'})
			//	-[r1:contains]-(node:Node {subsystem: 'cluster', type: 'node'})
			//	-[s1:contains]-(io:Node {subsystem: 'io'})
			// WHERE io.type = 'shm'
			// MATCH (node) -[r2:contains]-(socket:Node {subsystem: 'cluster', type: 'socket'})
			//	-[r3:contains]-(core:Node {subsystem: 'cluster', type: 'core'})
			// RETURN *
			// If we have a last scene, it needs to be a new MATCH
			newQuery := fmt.Sprintf("-[%s:contains]-(%s:Node {subsystem: 'cluster', type: '%s'})", resourceEdgeName, resourceType, resourceType)
			if lastSeen != "" {
				newQuery = fmt.Sprintf("\nMATCH (%s) %s", lastSeen, newQuery)
			} else {
				newQuery = "\n" + newQuery
			}
			query += newQuery

			// If we hit a subsystem, we need to define last seen, because it will start a new
			// MATCH expression for the next time.
			if exists {
				query += matcher.GenerateCypher(&subsystemNeeds)
				lastSeen = resourceType
			} else {
				lastSeen = ""
			}
		}

		// Keep going until we find a slot. When replicas != 0
		// we have found a slot.
		if resource.Replicas != 0 {
			if resource.With != nil {
				for _, with := range resource.With {

					// Child functions should start at resources we haven't parsed yet
					// We need the last seen type to add after a WHERE
					lastSeen = ""
					if len(seenTypes) > 0 {
						lastSeen = seenTypes[len(seenTypes)-1]
					}

					// These are remaining resources we need to parse over
					remainingResources := utils.Diff(resourceTypes, seenTypes)
					err := updateQuery(with, remainingResources, lastSeen)
					if err != nil {
						return err
					}
				}
			}
		}
		return nil
	}

	// Then get results for that, and we need to pass in a scecond query
	// Right now do a query for each schedulable slot
	// Not sure if these can be combined into one
	for _, resource := range resources {

		// We need to go through the structure of the graph. If
		// there are no subsystem needs, we match all. Otherwise
		// we also look for an edge to the subsytem
		resourceTypes := []string{"rack", "node", "socket", "core"}

		query = "MATCH (cluster:Node {subsystem: 'cluster', type: 'cluster'})"
		updateQuery(resource, resourceTypes, "")

		// When we get here, we are at a slot, and can just add to the query the
		// requirements of counts
		totals := graph.ExtractResourceSlots(jobspec)

		// This is the query I'm going for now - not sure if entirely correct
		// Maybe someone can help me more on these some day when they have bandwidth
		// MATCH (cluster:Node {subsystem: 'cluster', type: 'cluster'})
		// -[rackEdge:contains]-(rack:Node {subsystem: 'cluster', type: 'rack'})
		// -[nodeEdge:contains]-(node:Node {subsystem: 'cluster', type: 'node'})
		// -[contains]-(io:Node {subsystem: 'io'})
		// WHERE io.type = 'shm'
		// MATCH (node) -[socketEdge:contains]-(socket:Node {subsystem: 'cluster', type: 'socket'})
		// -[coreEdge:contains]-(core:Node {subsystem: 'cluster', type: 'core'})
		// WITH node,count(coreEdge) as cores_count
		// WHERE cores_count >= 3
		// RETURN node, cores_count

		for _, slotCount := range totals {
			query += fmt.Sprintf("\nWITH cluster,%s,count(distinct %sEdge) as %s_count", slotCount.Parent, slotCount.Name, slotCount.Name)
			query += fmt.Sprintf("\nWHERE %s_count >= %d", slotCount.Name, slotCount.Members)
		}

		// This assumes the return statement is the highest level of the slot
		topSlot := totals[0]
		query += fmt.Sprintf("\nRETURN cluster,%s, %s_count", topSlot.Parent, topSlot.Name)
		fmt.Printf("\n%s\n", query)
	}

	// Connect to the driver
	driver, err := neo4j.NewDriverWithContext(memoryHost, neo4j.BasicAuth(username, password, databaseName))
	if err != nil {
		return matches, err
	}
	ctx := context.Background()
	defer driver.Close(ctx)
	err = driver.VerifyConnectivity(ctx)
	if err != nil {
		return matches, err
	}

	// Do the query
	result, err := neo4j.ExecuteQuery(ctx, driver, query, nil, neo4j.EagerResultTransformer, neo4j.ExecuteQueryWithDatabase(databaseName))
	if err != nil {
		return matches, err
	}

	// Keep a count of matches per cluster
	lookup := map[string]int32{}

	// Print the node results
	for _, node := range result.Records {
		// Here is how to inspect additional node metadata
		// fmt.Println(node.AsMap()["cluster"].(neo4j.Node))                 // Node type
		// fmt.Println(node.AsMap()["cluster"].(neo4j.Node).GetProperties()) // Node properties
		// fmt.Println(node.AsMap()["cluster"].(neo4j.Node).GetElementId())  // Node internal ID
		// fmt.Println(node.AsMap()["cluster"].(neo4j.Node).Labels)          // Node labels
		clusterName := node.AsMap()["cluster"].(neo4j.Node).Props["name"].(string)

		// This gets rid of the prefix
		originalName := strings.Replace(clusterName, "cluster-", "", 1)

		// And the suffix
		parts := strings.Split(originalName, "-")
		originalName = strings.Join(parts[0:len(parts)-1], "-")
		_, ok := lookup[originalName]
		if !ok {
			lookup[originalName] = 0
		}
		lookup[originalName] += 1
	}

	// Keep matches that we have minimum slot count
	for cluster := range lookup {
		matches = append(matches, cluster)
	}
	fmt.Printf("\nMatches: %s\n", matches)
	return matches, nil
}

// Init provides extra initialization functionality
// We check credentials here
func (g Memgraph) Init(
	options map[string]string,
) error {

	// Warning: this assumes one client running with one graph host
	host, ok := options["memoryHost"]
	if ok {
		memoryHost = host
	}
	user, ok := options["username"]
	if ok {
		username = user
	}
	pw, ok := options["password"]
	if ok {
		password = pw
	}
	return nil
}

// Add the backend to be known to rainbow
func init() {

	graph := Memgraph{}
	backend.Register(graph)
}
