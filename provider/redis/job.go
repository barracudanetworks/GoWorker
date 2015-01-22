package redis

import "github.com/barracudanetworks/GoWorker/job"

// RedisJob contains information about a job provided by redis
type RedisJob struct {
	config   *job.JobConfig
	provider *Redis
}

// Config return the JobConfig for this job
func (r *RedisJob) Config() *job.JobConfig {
	return r.config
}

// JobConfirmer return this job's provider
func (r *RedisJob) JobConfirmer() job.JobConfirmer {
	return r.provider
}
