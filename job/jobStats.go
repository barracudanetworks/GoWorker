package job

import (
	"fmt"
	"time"
)

const (
	STATUS_NEW     Status = 0
	STATUS_STARTED Status = 1
	STATUS_SUCCESS Status = 2
	STATUS_FAILURE Status = 3
	STATUS_RETRY   Status = 4
)

type Status uint8

// JobStats holds information about a run of a job
type JobStats struct {
	startTime time.Time
	endTime   time.Time
	retries   int
	status    Status
}

// Start signals the start of a job
func (j *JobStats) Start() {
	j.startTime = time.Now()
	j.status = STATUS_STARTED
}

// End signals the end of a job
func (j *JobStats) End(s Status) {
	j.status = s
	j.endTime = time.Now()
}

// Retry signal the restart of a job
func (j *JobStats) Retry() {
	j.retries += 1
	j.status = STATUS_RETRY
}

// Retries returns the number of retries a job took to complete
func (j *JobStats) Retries() int {
	return j.retries

}

// Status returns the status of the job
func (j *JobStats) Status() Status {
	return j.status
}

// Duration return how long the job took to complete
func (j *JobStats) Duration() time.Duration {
	return j.endTime.Sub(j.startTime)
}

// String return the string representation of the jobStats object
func (j *JobStats) String() string {
	return fmt.Sprint(*j)
}

// NewJobStats initilizes a JobStats object
func NewJobStats() *JobStats {
	j := &JobStats{}
	j.Start()
	return j
}
