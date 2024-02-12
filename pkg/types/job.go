package types

// JobSpec holds basic metadata about a job
type JobSpec struct {
	Name    string
	Command string
	Nodes   int32
	Tasks   int32
}
