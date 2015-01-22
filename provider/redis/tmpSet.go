package redis

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"hash"
	"log"
	"time"

	"sync"

	"github.com/barracudanetworks/GoWorker/job"
	"github.com/barracudanetworks/GoWorker/lua"
	redigo "github.com/garyburd/redigo/redis"
)

var (
	JOB_NOT_FOUND = errors.New("redis: job not found")
)

const (
	TMP_JOB_LOCK_PREFIX = "tmp_job:lock:"
)

// TmpSet is a set in redis which holds temporary data
type TmpSet struct {
	fuzzyGet   *redigo.Script
	getOrphan  *redigo.Script
	popAndLock *redigo.Script
	confirm    *redigo.Script
	locks      map[*RedisJob]*keepAlive
	sha1       hash.Hash
	*sync.Mutex
	prefix string
}

// Get a single job and lock it
func (t *TmpSet) PopAndLock(r *Redis) (*RedisJob, error) {
	r.Lock()
	iJob, err := t.popAndLock.Do(r.conn, r.JobList, 30)
	raw, rErr := redigo.Bytes(iJob, err)
	if rErr != nil {
		r.Unlock()
		return nil, rErr
	}

	// let go of the connections
	r.Unlock()

	jobConfig, err := job.ParseConfig(raw)
	if err != nil {
		return nil, err
	}

	job := &RedisJob{
		config:   jobConfig,
		provider: r,
	}

	t.Lock()
	// rest the hasher
	t.sha1.Reset()
	t.sha1.Write(raw)

	// set the lock key
	keep := &keepAlive{
		killChan: make(chan struct{}),
		ttl:      time.Duration(30),
		key:      fmt.Sprintf("%x", t.sha1.Sum(nil)),
		job:      job,
	}
	log.Println(keep.key)
	t.sha1.Reset()

	// start the keep alive
	go keep.KeepAlive(r)

	t.locks[job] = keep
	t.Unlock()

	return job, nil
}

// ConfirmJob confirms a job
func (t *TmpSet) ConfirmJob(j *RedisJob, r *Redis) error {
	t.Lock()
	defer t.Unlock()
	k, ok := t.locks[j]
	if !ok {
		return JOB_NOT_FOUND
	}

	// stop updating the job's lock
	k.Kill()
	r.Lock()
	t.confirm.Do(r.conn, k.key)
	r.Unlock()

	// delete the lock
	delete(t.locks, j)

	return nil
}

// GetAllOrphan gets all of the orphaned jobs in the redis list
func (t *TmpSet) GetOrphan(r *Redis, max int) ([]*RedisJob, error) {
	r.Lock()
	defer r.Unlock()

	tmp, err := t.getOrphan.Do(r.conn, max)
	rawJobs, bErr := redigo.Strings(tmp, err)
	if bErr != nil {
		log.Println(bErr)
		return []*RedisJob{}, err
	}

	// make a slice to hold all of the jobs once they are parsed

	jobs := make([]*RedisJob, len(rawJobs))
	for i := 0; i < len(jobs); i++ {

		// parse the json blob
		conf, pErr := job.ParseConfig([]byte(rawJobs[i]))
		if pErr != nil {
			return jobs, pErr
		}

		jobs[i] = &RedisJob{
			provider: r,
			config:   conf,
		}
		t.Lock()
		// rest the hasher
		t.sha1.Reset()
		t.sha1.Write([]byte(rawJobs[i]))

		// set the lock key
		keep := &keepAlive{
			killChan: make(chan struct{}),
			ttl:      time.Duration(30),
			key:      fmt.Sprintf("%x", t.sha1.Sum(nil)),
			job:      jobs[i],
		}
		t.sha1.Reset()

		// start the keep alive
		go keep.KeepAlive(r)

		t.locks[jobs[i]] = keep

		t.Unlock()
	}

	return jobs, nil
}

// NewTmpSet create and return a new TmpSet
func NewTmpSet(prefix string) *TmpSet {
	t := &TmpSet{
		fuzzyGet:   lua.GET_BY_FUZZY_KEY_SCRIPT,
		getOrphan:  lua.GET_ORPHAN_SCRIPT,
		popAndLock: lua.POP_AND_LOCK_SCRIPT,
		confirm:    lua.CONFIRM_SCRIPT,
		prefix:     prefix,
		locks:      make(map[*RedisJob]*keepAlive),
		Mutex:      &sync.Mutex{},
		sha1:       sha1.New(),
	}
	return t
}
