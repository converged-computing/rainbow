package database

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	pb "github.com/converged-computing/rainbow/pkg/api/v1"
	_ "github.com/mattn/go-sqlite3"
)

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

// addJob adds a job to the jobs table
func (db *Database) addJob(job *pb.SubmitJobRequest) (*Job, error) {

	j := Job{}
	conn, err := db.connect()
	if err != nil {
		return &j, err
	}
	defer conn.Close()

	// The jobspec is added once to the database, first without assignment
	fields := "(name, jobspec)"
	values := fmt.Sprintf("(\"%s\", \"%s\")", job.Name, job.Jobspec)

	// Submit the query to get the global id (jobid, not submit yet)
	query := fmt.Sprintf("INSERT into jobs %s VALUES %s", fields, values)

	// Since we want to get a result back, we use query
	statement, err := conn.Prepare(query)
	if err != nil {
		return &j, err
	}
	defer statement.Close()

	// We expect only one job
	rows, err := statement.Query()
	if err != nil {
		return &j, err
	}

	// Unwrap into job
	for rows.Next() {
		err := rows.Scan(&j.Id, &j.Cluster, &j.Name, &j.Nodes, &j.Tasks, &j.Command)
		if err != nil {
			return &j, err
		}
	}
	return &j, nil
}

// addAssignments add jobs and clsuters (not assigned) to the assignment database
func (db *Database) addAssignments(job *Job, clusters []*Cluster) error {
	conn, err := db.connect()
	if err != nil {
		return err
	}
	defer conn.Close()

	// The jobspec is added once to the database, first without assignment
	fields := "(cluster, job, status)"

	// Assemble values, one set for each cluster
	values := ""
	for _, cluster := range clusters {
		if values == "" {
			// Status of 0 means unassigned (we don't know)
			values = fmt.Sprintf("(\"%s\", \"%d\", \"%d\")", cluster.Name, job.Id, 0)
		} else {
			values += fmt.Sprintf(",(\"%s\", \"%d\",\"%d\")", cluster.Name, job.Id, 0)
		}
	}

	// Submit the query to get the global id (jobid, not submit yet)
	query := fmt.Sprintf("INSERT into assign %s VALUES %s", fields, values)

	// Since we want to get a result back, we use query
	statement, err := conn.Prepare(query)
	if err != nil {
		return err
	}
	defer statement.Close()
	_, err = statement.Query()
	return err
}

// SubmitJob adds the assigned job to the database
func (db *Database) SubmitJob(
	job *pb.SubmitJobRequest,
	cluster *Cluster,
) (*pb.SubmitJobResponse, error) {

	response := &pb.SubmitJobResponse{}

	// Add the job to the database
	// TODO: should we do a check to see if we have the job already?
	// could create a hash / use the jobspec. Do we allow that?
	j, err := db.addJob(job)
	if err != nil {
		response.Status = pb.SubmitJobResponse_SUBMIT_ERROR
		return response, err
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

	// If the max jobs is < 1, we are asking to see all jobs
	query := fmt.Sprintf("SELECT * FROM jobs WHERE cluster = '%s'", cluster.Name)
	if request.MaxJobs >= 1 {
		query = fmt.Sprintf("%s LIMIT %d", query, request.MaxJobs)
	}

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

// AcceptJobs
// We use this function after validating a cluster with a secret
// and simply retrieve the ids and delete them from the database if they exist
func (db *Database) AcceptJobs(
	request *pb.AcceptJobsRequest,
	cluster *Cluster,
) (*pb.AcceptJobsResponse, error) {

	response := &pb.AcceptJobsResponse{}
	conn, err := db.connect()
	if err != nil {
		return response, err
	}
	defer conn.Close()

	// Select up to the limit of jobs
	jobids := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(request.Jobids)), ","), "[]")
	query := fmt.Sprintf("DELETE FROM jobs WHERE cluster = '%s' AND idJob in (%s)", cluster.Name, jobids)
	result, err := conn.Exec(query)

	// Error with request
	if err != nil {
		response.Status = pb.AcceptJobsResponse_RESULT_TYPE_ERROR
		return response, err
	}
	count, err := result.RowsAffected()
	log.Printf("%s: (%d)\n", query, count)

	response.Status = pb.AcceptJobsResponse_RESULT_TYPE_PARTIAL
	if count == int64(len(request.Jobids)) {
		response.Status = pb.AcceptJobsResponse_RESULT_TYPE_SUCCESS
	}
	return response, err
}
