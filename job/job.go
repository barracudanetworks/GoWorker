package job

import "errors"

var (
	JOB_NOT_FOUND = errors.New("job: job not found")
)

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

// BaseJob a simple job
type BaseJob struct {
	provider JobConfirmer
	raw      []byte
	conf     *JobConfig
}

// Config return the job config
func (b *BaseJob) Config() *JobConfig {
	return b.conf
}

// JobConfirmer return the provider this job points back at
func (b *BaseJob) JobConfirmer() JobConfirmer {
	return b.provider
}

func NewBaseJob(p JobConfirmer, conf *JobConfig, raw []byte) Job {
	return &BaseJob{
		provider: p,
		conf:     conf,
		raw:      raw,
	}
}
