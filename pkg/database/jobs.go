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
	Jobspec string `json:"jobspec"`
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
func (db *Database) addJob(job *pb.SubmitJobRequest, cluster string) (*Job, error) {

	j := Job{}
	conn, err := db.connect()
	if err != nil {
		return &j, err
	}
	defer conn.Close()

	// The jobspec is added once to the database, first without assignment
	fields := "(name, cluster, jobspec)"
	values := fmt.Sprintf("(\"%s\", \"%s\", \"%s\")", job.Name, cluster, job.Jobspec)

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
		err := rows.Scan(&j.Id, &j.Cluster, &j.Name, &j.Jobspec)
		if err != nil {
			return &j, err
		}
	}
	return &j, nil
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
	j, err := db.addJob(job, cluster.Name)
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
func (db *Database) ReceiveJobs(
	request *pb.ReceiveJobsRequest,
	cluster *Cluster,
) (*pb.ReceiveJobsResponse, error) {

	response := &pb.ReceiveJobsResponse{}
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
	response.Status = pb.ReceiveJobsResponse_REQUEST_JOBS_ERROR

	// Unwrap into list of jobs
	jobs := map[int32]string{}
	var j Job
	for rows.Next() {
		err := rows.Scan(&j.Id, &j.Cluster, &j.Name, &j.Jobspec)
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
		response.Status = pb.ReceiveJobsResponse_REQUEST_JOBS_NORESULTS
	} else {
		response.Status = pb.ReceiveJobsResponse_REQUEST_JOBS_SUCCESS
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
