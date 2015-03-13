package manager

import (
	"time"

	"github.com/eliothedeman/workhorse/job"
	"github.com/eliothedeman/workhorse/provider"
	"github.com/eliothedeman/workhorse/worker"
)

// JobTracker keeps track of jobs and the workers and providers they relate to
type JobTracker struct {
	jobToPack map[job.Job]*trackPack
	input     chan *trackPack
	remove    chan job.Job
	age       chan *agePack
	stop      chan struct{}
}

// NewJobTracker create and return a new JobTracker
func NewJobTracker() *JobTracker {
	return &JobTracker{
		jobToPack: make(map[job.Job]*trackPack),
		input:     make(chan *trackPack),
		remove:    make(chan job.Job),
		stop:      make(chan struct{}),
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
	var tmpAge *agePack
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
			for _, tmpPack = range j.jobToPack {
				j.unTrackJob(tmpPack.job)
			}

		// get all of the jobs created before the give time
		case tmpAge = <-j.age:
			tmpAge.r <- j.byAge(tmpAge)

		}

	}
}

func (j *JobTracker) Stop() {
	j.stop <- struct{}{}
}

// byAge return all jobs that are at least "d" old
func (j *JobTracker) byAge(age *agePack) []*trackPack {
	pack := make([]*trackPack, 0, 8)
	for _, v := range j.jobToPack {
		if v.startTime.Before(age.age) {
			pack = append(pack, v)
		}
	}
	return pack
}

// ByAge
func (j *JobTracker) ByAge(age time.Time) []*trackPack {
	ap := &agePack{
		age: age,
		r:   make(chan []*trackPack),
	}
	j.age <- ap

	return <-ap.r

}

func (j *JobTracker) trackJob(p *trackPack) {
	j.jobToPack[p.job] = p
}

func (j *JobTracker) unTrackJob(jb job.Job) {
	delete(j.jobToPack, jb)
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
	id        uint64
	job       job.Job
	worker    worker.Worker
	provider  provider.Provider
	startTime time.Time
}

type agePack struct {
	age time.Time
	r   chan []*trackPack
}
