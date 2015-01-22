/*
Package redis contains a provider which supplies the manager with jobs stored in a redis list.
*/
package redis

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/barracudanetworks/GoWorker/job"
	"github.com/barracudanetworks/GoWorker/provider"
	"github.com/barracudanetworks/GoWorker/worker/disk"
	"github.com/eliothedeman/memString"
	redigo "github.com/garyburd/redigo/redis"
)

const (
	DEFAULT_POOL_SIZE = 2
	DEFAULT_MAX_IDLE  = 100
	DEFAULT_HOST      = "localhost"
	DEFAULT_PORT      = "6379"
	DEFAULT_JOB_LIST  = "job_list"
	TEMP_JOB_LIST     = "tmp_job_list"
)

var (
	MAX_WAIT_TIME    = 10 * time.Second
	DEFAULT_TIMEOUT  = 10 * time.Second
	REDIS_INFO_ERROR = errors.New("redis: failed to parse redis info")
)

// init load this provider into the master map of providers
func init() {
	if provider.Factories != nil {
		provider.Factories["redis"] = RedisFactory
	} else {
		log.Println("Unable to load redis provider factory")
	}
}

// RedisConfig contains config options for a redis provider
type RedisConfig struct {
	Host        string  `json:"host" required:"true" description:"The host of the redis server to connect to"`
	Port        string  `json:"port" required:"true" description:"Port of the redis server to connect to."`
	JobList     string  `json:"job_list" required:"true" description:"The list in redis to pull jobs from."`
	DumpOnLimit bool    `json:"dump_on_limit" required:"false" description:"When the redis server reaches this level of memory, start dumping the job list to disk. The file worker must be enabled to use this feature."`
	MemLimit    string  `json:"memory_limit" required:"false" description:"The point at which to dump the job list to disk. This will have no effect if dump_on_limit is not enabled."`
	Target      float64 `json:"target" required:"false" description:"The target jobs per second for this jobs on this job_list."`
}

// Redis holds a pool of redis connections that are used to talk to the database
type Redis struct {
	conn redigo.Conn
	*sync.Mutex
	host        string
	port        string
	tmpSet      *TmpSet
	JobList     string
	memoryLimit int64
	dumpOnLImit bool
	lastJobChan chan job.Job
	target      float64
}

func (r *Redis) Target() float64 {
	return r.target
}

func (r *Redis) ConfigStruct() interface{} {
	return &RedisConfig{}
}

// Init initilize the redis provider
func (r *Redis) Init(i interface{}) error {
	conf, ok := i.(*RedisConfig)
	if !ok {
		return provider.WRONG_CONFIG_TYPE
	}
	c, err := redigo.Dial("tcp", conf.Host+":"+conf.Port)
	r.host = conf.Host
	r.port = conf.Port
	r.conn = c
	r.Mutex = &sync.Mutex{}
	r.JobList = conf.JobList
	r.tmpSet = NewTmpSet("tmp_job:")
	r.dumpOnLImit = conf.DumpOnLimit
	r.memoryLimit, _ = memString.ParseMemory(conf.MemLimit)
	r.target = conf.Target

	// if we need to dump on memory limit, start a routine to check for the limit
	if r.dumpOnLImit {
		go func(m int64) {
			for {
				if r.CheckMemory(m) {
					log.Println("Memory limit of", m, "bytes has been reached on redis hosts", r.host+":"+r.port)
					stop := make(chan struct{})
					// drain the list until we are under the limit
					go r.Drain(stop)
					for {
						// if we are under the limit, stop draining
						if !r.CheckMemory(m) {
							stop <- struct{}{}
							break
						}
						time.Sleep(1 * time.Second)
					}
				}
				time.Sleep(10 * time.Second)
			}
		}(r.memoryLimit)
	}
	return err
}

// Drain pull jobs off redis and wrap them in a file job to be written to disk.
// this can be done if redis starts running out of memory.
// Drain will continue to remove jobs from redis until it either:
// a : gets a signal on it's stop channel
// b : runs out of jobs in the redis cue
// c : the connection to redis is closed
func (r *Redis) Drain(stopChan chan struct{}) {
	r.Lock()
	jobChan := r.lastJobChan
	r.Unlock()
	log.Println("Started draining jobs from list", r.JobList)
	var err error
	for {
		select {
		case <-stopChan:
			break
		default:
			j := r.popJob()
			if j == nil {
				return
			}
			conf := j.Config()
			// wrap the job in a file job
			fc := &disk.DiskParams{}
			fc.ExicutionTime = time.Now().Unix()
			fc.Job, err = json.Marshal(conf)
			if err != nil {
				log.Println(err)
				continue
			}
			// switch the job to a file type
			conf.Type = "file"
			var b []byte
			b, err = json.Marshal(fc)
			if err != nil {
				log.Println(err)
			}
			conf.Params = json.RawMessage(b)
			jobChan <- j
		}
	}
	log.Println("Done draining jobs fro list", r.JobList)
}

// CheckMemory check to see if the memory limit for the redis server has been hit
func (r *Redis) CheckMemory(max int64) bool {
	var b []byte
	var err error
	var i interface{}
	r.Lock()
	i, err = r.conn.Do("info")
	r.Unlock()
	b, err = redigo.Bytes(i, err)
	if err != nil {
		log.Println(err)
		return false
	}
	n, err := parseRedisInfoInt(b, "used_memory")
	return n > r.memoryLimit
}

// popJob pops a job off of the redis list then pushes it to the temparary list
func (r *Redis) popJob() job.Job {
	job, err := r.tmpSet.PopAndLock(r)
	if err != nil {
		log.Println(err)
		return nil
	}
	return job
}

// lenList get the length of a give list
func (r *Redis) lenList(list string) uint64 {
	r.Lock()
	defer r.Unlock()
	var l uint64
	v, err := r.conn.Do("llen", list)
	l, err = redigo.Uint64(v, err)
	if err != nil {
		log.Println(err)
		return 0
	}
	return l
}

// pushJob pushes a job onto the redis list
func (r *Redis) pushJob(j *RedisJob, list string) error {
	b, err := json.Marshal(j.config)
	if err != nil {
		return err
	}
	r.Lock()
	defer r.Unlock()
	_, err = r.conn.Do("lpush", list, b)
	return err
}

// ConfirmJob removes the job from the tmp list on the redis server, signifying success
func (r *Redis) ConfirmJob(j job.Job) error {
	r.tmpSet.ConfirmJob(j.(*RedisJob), r)
	return nil
}

// createJob returns a pointer to a new RedisJob
func (r *Redis) createJob(config *job.JobConfig) *RedisJob {
	return &RedisJob{
		config:   config,
		provider: r,
	}
}

// Close close all of the connections to redis
func (r *Redis) Close() error {
	r.conn.Close()
	return r.conn.Close()
}

// Name
func (r *Redis) Name() string {
	return "redis_" + r.JobList
}

// RequestWork request parse jobs and send them to the manager to be processed
func (r *Redis) RequestWork(num int, jobChan chan job.Job) error {
	r.Lock()
	r.lastJobChan = jobChan
	r.Unlock()
	// get orphan jobs first
	orphans, err := r.tmpSet.GetOrphan(r, num)
	if err != nil {
		return err
	}
	for i := 0; i < len(orphans); i++ {
		jobChan <- orphans[i]
	}
	num -= len(orphans)

	// only attempt to get as many jobs as are in the redis queue
	numJobs := r.lenList(r.JobList)
	if int(numJobs) < num {
		num = int(numJobs)
	}

	// get new jobs that have not been orphaned
	var j job.Job
	for i := 0; i < num; i++ {
		j = r.popJob()
		if j != nil {
			jobChan <- j
		}
	}
	return nil
}

// WaitTime given a target of jobs persecond, how long should the manager wait before asking for more work
func (r *Redis) WaitTime(target float64) time.Duration {
	return 10 * time.Second
}

// NewRedis create a new redis connection provider
func NewRedis(url string, poolSize int, JobList string) (*Redis, error) {
	c, err := redigo.Dial("tcp", url)
	r := &Redis{
		conn:    c,
		Mutex:   &sync.Mutex{},
		JobList: JobList,
		tmpSet:  NewTmpSet("test_prefix:"),
	}
	return r, err
}

// RedisFactory constructs a new redis provider from a ProviderConfig and returns it along with any errors
func RedisFactory() provider.Provider {
	return &Redis{}
}

// parseRedisInfoInt parse the response from a redis info call into an int field for the given field
func parseRedisInfoInt(b []byte, key string) (int64, error) {
	buff := bytes.NewBuffer(b)
	var err error
	var i int64
	scanner := bufio.NewScanner(buff)
	for scanner.Scan() {
		h := scanner.Bytes()
		if bytes.HasPrefix(h, []byte(key)) {
			s := bytes.Split(h, []byte(":"))
			if len(s) < 2 {
				return i, REDIS_INFO_ERROR
			}
			i, err = memString.ParseMemory(string(s[1]))
			if err != nil {
				return i, err
			}
			break
		}
	}
	return i, err
}
