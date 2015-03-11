package manager

import (
	"github.com/eliothedeman/img/provider"
	"github.com/eliothedeman/workhorse/job"
	"github.com/eliothedeman/workhorse/worker"
)

// JobTracker keeps track of jobs and the workers and providers they relate to
type JobTracker struct {
	idToJob      map[int]*job.Job
	idToWorker   map[int]*worker.Worker
	idToProvider map[int]*provider.Provider
	jobToId      map[*job.Job]int
	input        chan TrackPack
	remove       chan job.Job
}

func (j *JobTracker) Track()

func (j *JobTracker) Add(id int, job job.Job, w worker.Worker, p provider.Provider) {
	jt := trackPack{
		id:       id,
		job:      job,
		worker:   w,
		provider: p,
	}
	j.input <- j
}

func (j *JobTracker) Remove(job job.job) {
	j.remove <- job
}

type trackPack struct {
	id       int
	job      job.Job
	worker   worker.Worker
	provider provider.Provider
}
