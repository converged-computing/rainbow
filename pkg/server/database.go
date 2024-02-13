package server

import (
	"encoding/json"
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
	Token  string
}

type Job struct {
	Id      int32  `json:"id"`
	Cluster string `json:"cluster"`
	Name    string `json:"name"`
	Nodes   int32  `json:"nodes"`
	Tasks   int32  `json:"tasks"`
	Command string `json:"command"`
}

// ToJson converts the job to json for sending back!
func (j *Job) ToJson() (string, error) {
	b, err := json.Marshal(j)
	if err != nil {
		return "", err
	}
	return string(b), nil
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
func (db *Database) RegisterCluster(name string) (*pb.RegisterResponse, error) {
	response := &pb.RegisterResponse{}

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
	// The "secret" is used by the cluster to request jobs for itself
	token := uuid.New().String()
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

// SubmitJob adds the job to the database
func (db *Database) SubmitJob(
	job *pb.SubmitJobRequest,
	cluster *Cluster,
) (*pb.SubmitJobResponse, error) {

	response := &pb.SubmitJobResponse{}
	conn, err := db.connect()
	if err != nil {
		return response, err
	}
	defer conn.Close()

	// Prepare the sql to insert the job
	fields := "(cluster, name, nodes, tasks, command)"
	values := fmt.Sprintf(
		"(\"%s\", \"%s\",\"%d\",\"%d\",\"%s\")",
		cluster.Name, job.Name, job.Nodes, job.Tasks, job.Command,
	)

	// Submit the query to get the global id (jobid, not submit yet)
	query := fmt.Sprintf("INSERT into jobs %s VALUES %s", fields, values)

	// From this point on (until the end) any early return is an error
	response.Status = pb.SubmitJobResponse_SUBMIT_ERROR

	// Since we want to get a result back, we use query
	statement, err := conn.Prepare(query)
	if err != nil {
		return response, err
	}
	defer statement.Close()

	// We expect only one job
	rows, err := statement.Query()
	if err != nil {
		return response, err
	}

	// Unwrap into job
	j := Job{}
	for rows.Next() {
		err := rows.Scan(&j.Id, &j.Cluster, &j.Name, &j.Nodes, &j.Tasks, &j.Command)
		if err != nil {
			return response, err
		}
	}

	// Success!
	response.Status = pb.SubmitJobResponse_SUBMIT_SUCCESS
	response.Jobid = j.Id
	return response, nil
}

// Request MaxJobs for a cluster to receive
func (db *Database) RequestJobs(
	request *pb.RequestJobsRequest,
	cluster *Cluster,
) (*pb.RequestJobsResponse, error) {

	response := &pb.RequestJobsResponse{}
	conn, err := db.connect()
	if err != nil {
		return response, err
	}
	defer conn.Close()

	// Select up to the limit of jobs
	query := fmt.Sprintf("SELECT * FROM jobs WHERE cluster = '%s' LIMIT %d", cluster.Name, request.MaxJobs)

	// Since we want to get a result back, we use query
	statement, err := conn.Prepare(query)
	if err != nil {
		return response, err
	}
	defer statement.Close()

	// We expect only one job
	rows, err := statement.Query()
	if err != nil {
		return response, err
	}

	// Failures from here until end are error
	response.Status = pb.RequestJobsResponse_REQUEST_JOBS_ERROR

	// Unwrap into list of jobs
	jobs := map[int32]string{}
	var j Job
	for rows.Next() {
		err := rows.Scan(&j.Id, &j.Cluster, &j.Name, &j.Nodes, &j.Tasks, &j.Command)
		if err != nil {
			return response, err
		}
		jobstr, err := j.ToJson()
		if err != nil {
			return response, err
		}
		jobs[j.Id] = jobstr
	}

	// No jobs, a quick check
	if len(jobs) == 0 {
		response.Status = pb.RequestJobsResponse_REQUEST_JOBS_NORESULTS
	} else {
		response.Status = pb.RequestJobsResponse_REQUEST_JOBS_SUCCESS
	}
	// Success! This is a lookup of job ids to the serialized string
	response.Jobs = jobs
	return response, nil
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

// GetCluster gets a cluster by name
func (db *Database) GetCluster(name string) (*Cluster, error) {

	// Connect!
	conn, err := db.connect()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// First determine if it exists
	query := fmt.Sprintf("SELECT * from clusters WHERE name LIKE \"%s\" LIMIT 1", name)
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
		err := rows.Scan(&cluster.Name, &cluster.Token, &cluster.Secret)
		if err != nil {
			return nil, err
		}
	}
	// Debugging extra for now
	log.Printf("%s: %s\n", query, cluster.Name)
	return &cluster, nil
}

// ValidateClusterToken checks if a cluster token is valid
// The token is used for validating a submission request.
func (db *Database) ValidateClusterToken(name, token string) (*Cluster, error) {
	cluster, err := db.GetCluster(name)
	if err != nil {
		return nil, err
	}

	// Validate the name and token
	if cluster.Name == "" || cluster.Token != token {
		return nil, fmt.Errorf("request denied")
	}
	return cluster, nil
}

// ValidateClusterSecret checks if a cluster secret is valid
// The secret is used for validating a request for jobs.
func (db *Database) ValidateClusterSecret(name, secret string) (*Cluster, error) {
	cluster, err := db.GetCluster(name)
	if err != nil {
		return nil, err
	}

	// Validate the name and token
	if cluster.Name == "" || cluster.Secret != secret {
		return nil, fmt.Errorf("request denied")
	}
	return cluster, nil
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
		token TEXT,
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
