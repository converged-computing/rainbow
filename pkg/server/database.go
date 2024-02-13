package server

import (
	"fmt"

	pb "github.com/converged-computing/rainbow/pkg/api/v1"
	"github.com/converged-computing/rainbow/pkg/utils"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"

	"database/sql"
	"log"
	"os"
)

type Database struct {
	filepath string
}

// Database types to serialize back into
type Cluster struct {
	Name   string
	Secret string
}

type Job struct {
	Id      int32
	Cluster string
	Name    string
	Nodes   int32
	Tasks   int32
	Command string
}

// cleanup removes the filepath
func (db *Database) cleanup() {
	// Delete a previous database that exists
	// Note that in the future we might not want to do this
	log.Printf("🧹️ cleaning up %s...", db.filepath)
	os.Remove(db.filepath)
}

// create the database
func (db *Database) create() error {

	log.Printf("✨️ creating %s...", db.filepath)

	// Create SQLite file (ensures that we can!)
	file, err := os.Create(db.filepath)
	if err != nil {
		return err
	}
	file.Close()
	log.Printf("   %s file created", db.filepath)

	// Open the created SQLite File (to test)
	conn, err := db.connect()
	defer conn.Close()
	return err
}

// Connect to the database - the caller is responsible for closing
func (db *Database) connect() (*sql.DB, error) {
	conn, err := sql.Open("sqlite3", db.filepath)
	if err != nil {
		return nil, err
	}
	return conn, err
}

// RegisterCluster registers a cluster or returns another status
// REGISTER_SUCCESS = 1;
// REGISTER_ERROR = 2;
// REGISTER_DENIED = 3;
// REGISTER_EXISTS = 4;
func (db *Database) RegisterCluster(name string) (pb.RegisterResponse_ResultType, string, error) {

	// Connect!
	conn, err := db.connect()
	if err != nil {
		return 0, "", err
	}
	defer conn.Close()

	// First determine if it exists - this needs to get the results
	query := fmt.Sprintf("SELECT count(*) from clusters WHERE name = '%s'", name)
	count, err := countResults(conn, query)
	if err != nil {
		return 0, "", err
	}
	// Debugging extra for now
	log.Printf("%s: (%d)\n", query, count)

	// Case 1: already exists
	if count > 0 {
		return 4, "", nil
	}

	// Generate a "secret" token, lol
	token := uuid.New().String()
	query = fmt.Sprintf("INSERT into clusters (name, secret) VALUES (\"%s\", \"%s\")", name, token)
	result, err := conn.Exec(query)
	if err != nil {
		return 2, "", err
	}
	count, err = result.RowsAffected()
	log.Printf("%s: (%d)\n", query, count)

	// REGISTER_SUCCESS
	if count > 0 {
		return 1, token, nil
	}

	// REGISTER_ERROR
	return 2, "", err
}

// SubmitJob adds the job to the database
// SUBMIT_UNSPECIFIED = 0;
// SUBMIT_SUCCESS = 1;
// SUBMIT_ERROR = 2;
// SUBMIT_DENIED = 3;
func (db *Database) SubmitJob(job *pb.SubmitJobRequest, cluster *Cluster) (pb.SubmitJobResponse_ResultType, int32, error) {
	var jobid int32
	conn, err := db.connect()
	if err != nil {
		return 0, jobid, err
	}
	defer conn.Close()

	// Generate a "secret" token, lol
	fields := "(cluster, name, nodes, tasks, command)"
	values := fmt.Sprintf("(\"%s\", \"%s\",\"%d\",\"%d\",\"%s\")", cluster.Name, job.Name, job.Nodes, job.Tasks, job.Command)

	// Submit the query to get the global id (jobid, not submit yet)
	query := fmt.Sprintf("INSERT into jobs %s VALUES %s", fields, values)

	// Since we want to get a result back, we use query
	statement, err := conn.Prepare(query)
	if err != nil {
		return 2, jobid, err
	}
	defer statement.Close()

	// We expect only one job
	rows, err := statement.Query()
	if err != nil {
		return 2, jobid, err
	}

	// Unwrap into job
	j := Job{}
	for rows.Next() {
		err := rows.Scan(&j.Id, &j.Cluster, &j.Name, &j.Nodes, &j.Tasks, &j.Command)
		if err != nil {
			return 2, jobid, err
		}
	}
	// Success
	return 1, j.Id, nil
}

// countResults counts the results for a specific query
func countResults(conn *sql.DB, query string) (int64, error) {

	var count int
	err := conn.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, err
	}
	return int64(count), nil
}

// GetCluster gets a cluster if it exists AND the token for it is valid
func (db *Database) GetCluster(name, token string) (*Cluster, error) {

	// Connect!
	conn, err := db.connect()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// First determine if it exists
	query := fmt.Sprintf("SELECT * from clusters WHERE name LIKE \"%s\" LIMIT 1", name)
	// Since we want to get a result back, we use query
	statement, err := conn.Prepare(query)
	if err != nil {
		return nil, err
	}

	// Only allow one result, one cluster
	rows, err := statement.Query()
	if err != nil {
		return nil, err
	}

	// Unwrap result into cluster
	cluster := Cluster{}
	for rows.Next() {
		err := rows.Scan(&cluster.Name, &cluster.Secret)
		if err != nil {
			return nil, err
		}
	}

	// Validate the name and token
	if cluster.Name == "" || cluster.Secret != token {
		return nil, fmt.Errorf("request denied")
	}
	// Debugging extra for now
	log.Printf("%s: %s\n", query, cluster.Name)
	return &cluster, nil
}

// create the database
func (db *Database) createTables() error {

	conn, err := db.connect()
	if err != nil {
		return err
	}
	defer conn.Close()

	// Create the clusters table, where we store the name and secret
	// obviously the secret should not be stored in plain text - it's fine for now
	createClusterTableSQL := `
	CREATE TABLE clusters (
		name TEXT NOT NULL PRIMARY KEY,		
		secret TEXT
	  );
	`

	createJobsTableSQL := `
	  CREATE TABLE jobs (
		  idJob integer NOT NULL PRIMARY KEY AUTOINCREMENT,		
		  cluster TEXT,
		  name TEXT,
		  nodes integer,
		  tasks integer,
		  command TEXT,
		  FOREIGN KEY(cluster) REFERENCES clusters(name)
		);`

	for table, statement := range map[string]string{"cluster": createClusterTableSQL, "jobs": createJobsTableSQL} {
		log.Printf("   create %s table...\n", table)
		query, err := conn.Prepare(statement) // Prepare SQL Statement
		if err != nil {
			return err
		}
		// Execute SQL query
		_, err = query.Exec()
		if err != nil {
			return err
		}
		log.Printf("   %s table created\n", table)
	}
	return nil
}

func initDatabase(filepath string, cleanup bool) (*Database, error) {

	// Create a new database (todo, add cleanupc check)
	db := Database{filepath: filepath}

	if cleanup {
		db.cleanup()
	}

	// If we haven't created yet or cleaned up
	exists, err := utils.PathExists(db.filepath)
	if err != nil {
		return nil, err
	}
	if !exists {
		// Create the database
		err := db.create()
		if err != nil {
			return nil, err
		}
		err = db.createTables()
	}
	return &db, err
}
