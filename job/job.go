package job

type JobType string

// Job contains all nessisary information to run a job on a worker
type Job interface {
	Config() *JobConfig
	JobConfirmer() JobConfirmer
}

// Confirms that a job has been completed
type JobConfirmer interface {
	ConfirmJob(j Job) error
}
