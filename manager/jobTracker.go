package manager

import (
	"time"

	"github.com/eliothedeman/workhorse/job"
	"github.com/eliothedeman/workhorse/provider"
	"github.com/eliothedeman/workhorse/worker"
)

// JobTracker keeps track of jobs and the workers and providers they relate to
type JobTracker struct {
	idToJob       map[uint64]job.Job
	idToWorker    map[uint64]worker.Worker
	idToProvider  map[uint64]provider.Provider
	idToStartTime map[uint64]time.Time
	jobToId       map[job.Job]uint64
	input         chan *trackPack
	remove        chan job.Job
	stop          chan struct{}
}

// NewJobTracker create and return a new JobTracker
func NewJobTracker() *JobTracker {
	return &JobTracker{
		idToJob:       make(map[uint64]job.Job),
		idToWorker:    make(map[uint64]worker.Worker),
		idToProvider:  make(map[uint64]provider.Provider),
		idToStartTime: make(map[uint64]time.Time),
		jobToId:       make(map[job.Job]uint64),
		input:         make(chan *trackPack),
		remove:        make(chan job.Job),
		stop:          make(chan struct{}),
	}
}

// NewAndStartTracker create and start a JobTracker
func NewAndStartTracker() *JobTracker {
	j := NewJobTracker()
	go j.Track()
	return j
}

// Track
func (j *JobTracker) Track() {
	var tmpPack *trackPack
	var tmpJob job.Job
	for {
		select {

		// start trackinga job
		case tmpPack = <-j.input:
			j.trackJob(tmpPack)

		// stop tracking a job
		case tmpJob = <-j.remove:
			j.unTrackJob(tmpJob)

		// stop tracking jobs
		case <-j.stop:
			for tmpJob, _ = range j.jobToId {
				j.unTrackJob(tmpJob)
			}
		}
	}
}

func (j *JobTracker) Stop() {
	j.stop <- struct{}{}
}

func (j *JobTracker) trackJob(p *trackPack) {
	j.idToJob[p.id] = p.job
	j.idToProvider[p.id] = p.provider
	j.idToWorker[p.id] = p.worker
	j.jobToId[p.job] = p.id
	j.idToStartTime[p.id] = time.Now()
}

func (j *JobTracker) unTrackJob(jb job.Job) {
	id, inMap := j.jobToId[jb]
	if inMap {
		delete(j.jobToId, jb)
		delete(j.idToJob, id)
		delete(j.idToProvider, id)
		delete(j.idToWorker, id)
	}
}

// Add start tracking a job
func (j *JobTracker) Add(id uint64, job job.Job, w worker.Worker, p provider.Provider) {
	jt := &trackPack{
		id:       id,
		job:      job,
		worker:   w,
		provider: p,
	}
	j.input <- jt
}

// Remove stop tracking a job
func (j *JobTracker) Remove(job job.Job) {
	j.remove <- job
}

type trackPack struct {
	id       uint64
	job      job.Job
	worker   worker.Worker
	provider provider.Provider
}
