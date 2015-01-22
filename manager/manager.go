package manager

import (
	"encoding/json"
	"log"
	"net/http"
	"runtime"
	"time"

	"github.com/garyburd/redigo/redis"

	"strings"

	"github.com/barracudanetworks/GoWorker/config"
	"github.com/barracudanetworks/GoWorker/job"
	"github.com/barracudanetworks/GoWorker/provider"
	"github.com/barracudanetworks/GoWorker/worker"
)

const (
	MANAGER_PREFIX          = "manager:"
	CONNECTION_REFRESH_RATE = 360
	DEFAULT_MAX_WORKERS     = 20
)

// Manager is used to manage worker processes
type Manager struct {
	// WorkerFactories a map of WorkerFactories that can create workers of a spesific type
	// WorkerFactories map[string]worker.WorkerFactory
	// // ProviderFactories a map of ProviderFactory taht can create providers fo a spesific type
	// ProviderFactories map[string]provider.ProviderFactory
	// KillChan if recieved on this channel, kill all workers
	KillChan chan struct{}
	// Stats keeps track of statistics of jobs running through the manager
	Stats *ManagerStats
	// Providers a map of provider.Provider's that can be used to request work
	Providers map[string]provider.Provider
	// AllWorkers, whether available or not
	allWorkers map[uint64]worker.Worker
	// readyWorkers a map to buffered channels of workers of various types
	readyWorkers map[string]chan worker.Worker
	// jobChan channel the manager uses to receive work
	jobChan chan job.Job
	// currentWorkers current number of workers per type
	currentWorkers map[string]int
	// redisConn is a connection to the redis server used for discovory
	redisConn redis.Conn
	// currentConfig holds the most recent app config
	currentConfig *config.AppConfig
	// statsServer
	statsServer *http.ServeMux
	// FailureHandler is a worker that handles failed jobs which have reached their retry limit
	failureHandlers []chan worker.Worker
	// handleFailures if this is set, the manager will look for a failure handler worker for failed jobs
	handleFailures bool
	numWorkers     *Counter
}

// Manage create and manage workers
func (m *Manager) Manage() {

	go func() {
		for {
			time.Sleep(10 * time.Second)
			runtime.GC()
		}
	}()
	// start the web server
	for n, p := range m.Providers {
		go func(n string, p provider.Provider) {
			// request ten jobs so we can see how long they take
			log.Printf("Requesting %d jobs from %s for load analysis", 10, n)
			m.RequestWork(p, 10)
			time.Sleep(5 * time.Second)
			for {
				count := m.numJobsToRequest(p, p.WaitTime(0))
				log.Printf("Requesting %d jobs from %s", count, n)
				m.RequestWork(p, count)
				time.Sleep(p.WaitTime(0))
			}
		}(n, p)
	}

	for {
		select {
		case <-m.KillChan:
			m.killAll()
			return
		case job := <-m.jobChan:
			m.runJob(job)
		}
	}
}

// numJobsToRequest calculate the number of jobs to request
func (m *Manager) numJobsToRequest(p provider.Provider, d time.Duration) int {

	// count = target jobs per second * time we have to do those jobs
	count := p.Target() * d.Seconds()

	// if we can't handle the target load, hard limit on our maximum
	if m.capasityByProvider(p) < p.Target() {
		count = m.capasityByProvider(p)
	}

	return int(count)
}

// capasityByProvider get the number of jobs we can handle from a provider.
// the result is given in jobs per second
func (m *Manager) capasityByProvider(p provider.Provider) float64 {

	// get the average duration it takes for work from this provider to complete
	avgDuration := m.Stats.AverageDurationByProvider(p).Seconds()

	// the capsity is defined by number of workers we have * jobs we can do per second per worker
	return float64(len(m.allWorkers)) * (time.Second.Seconds() / avgDuration)
}

// requestWork takes a map of providers, and request work from each of them
func (m *Manager) RequestWork(p provider.Provider, numJobs int) {
	p.RequestWork(numJobs, m.jobChan)
}

// runJob takes a job, finds the best suited worker, and runs the job and waits for the results, after which performs cleanup tasks
func (m *Manager) runJob(j job.Job) {
	config := j.Config()
	workerChan, ok := m.readyWorkers[config.Type]
	if !ok {
		log.Println("Unknown job type", config.Type)
		return
	}

	// attempt to get an existing worker, if non is available, and the worker limit has not been reached, create a new one
	worker := <-workerChan
	go func() {
		stats := worker.Work(j)

		// recycle worker and send it back on the worker chan
		worker.Recycle()
		workerChan <- worker

		// Log outcome of job
		log.Printf("%s completed with status %d and %d retries. Job took %s to complete", config.Name, stats.Status(), stats.Retries(), stats.Duration())

		// if their was a failure, set the job as a retry
		if stats.Status() != job.STATUS_SUCCESS {
			m.handleFailure(j, stats)
		} else {
			// if the job succeded confirm and consume stats normally
			if err := j.JobConfirmer().ConfirmJob(j); err != nil {
				log.Println(err)
			}
			m.Stats.consumeStats(j, stats)
		}
	}()
}

// HandleFailure given a job that has failed to complete, asses it's status and handle accordingly
func (m *Manager) handleFailure(j job.Job, s *job.JobStats) {

	config := j.Config()
	// if the job was not successful decrement it's retries
	config.Retries = config.Retries - 1

	// if it is out of retires and the manager has a way to handle the error, send it to the failure handler
	if config.Retries <= 0 {
		s.End(job.STATUS_FAILURE)
		m.Stats.consumeStats(j, s)
		if m.handleFailures {
			for i := range m.failureHandlers {
				worker := <-m.failureHandlers[i]
				log.Println("sending failed job to failure handler")
				stats := worker.Work(j)
				log.Printf("FAILURE_HANDLER::%s completed with status %d and %d retries. Job took %s to complete", config.Name, stats.Status(), stats.Retries(), stats.Duration())
				m.failureHandlers[i] <- worker
			}

		} else {
			log.Printf("job %+v failed with no failure handler provided\n", j)
		}
	} else {
		// set the job stats to a retry
		s.End(job.STATUS_RETRY)
		m.Stats.consumeStats(j, s)
		// if the job still has retries left, send it back to the manager to try again
		m.jobChan <- j
	}
	if err := j.JobConfirmer().ConfirmJob(j); err != nil {
		log.Println(err)
	}
}

// killAll kill all available workers, and set them to nil
func (m *Manager) killAll() error {
	var err error
	for _, worker := range m.allWorkers {
		worker.Kill()
		if err != nil {
			return err
		}
	}
	return nil
}

// applyConfig apply a new configuration to the manager
func (m *Manager) applyConfig(conf *config.AppConfig) {
	m.populateProviders(conf.ProviderConfigs)
	m.populateWorkers(conf.WorkerConfigs)
	m.populateFailureHandlers(conf.FailureHanldlerConfigs)
}

// populateProviders initilizes all providers supplied by the AppConfig
func (m *Manager) populateProviders(confs []config.ConfigPair) {
	for i := range confs {
		c := confs[i]
		p, err := provider.Create(c.Type, c.Config)
		if err != nil {
			log.Println(err)
			continue
		}
		m.Providers[c.Type] = p
	}
}

// populateWorkers creates a cache of workers for every type we have and puts it on the channel for that type
func (m *Manager) populateWorkers(confs []config.ConfigPair) {
	for i := range confs {
		c := confs[i]
		count := struct {
			NumWorkers int `json:"workers"`
		}{
			DEFAULT_MAX_WORKERS,
		}
		jErr := json.Unmarshal(c.Config, &count)
		if jErr != nil {
			count.NumWorkers = DEFAULT_MAX_WORKERS
		}

		// populate the workers
		m.readyWorkers[string(c.Type)] = make(chan worker.Worker, count.NumWorkers)
		for i := 0; i < count.NumWorkers; i++ {
			w, err := worker.Create(string(c.Type), c.Config)
			if err != nil {
				log.Println(err)
				break
			}
			m.readyWorkers[string(c.Type)] <- w
			m.allWorkers[m.numWorkers.Val()] = w
			m.numWorkers.Inc()
		}
		log.Println("created", count.NumWorkers, c.Type, "workers")
	}
}

// populateFialureHandler creates a worker to handle failed jobs which have reached their retry limit
func (m *Manager) populateFailureHandlers(confs []config.ConfigPair) {
	m.failureHandlers = make([]chan worker.Worker, 0)

	// for every config, add another error handler
	for i := range confs {
		c := confs[i]
		w, err := worker.Create(string(c.Type), c.Config)
		if err != nil {
			log.Println(err)
			log.Println("Unable to initilize failure handler")
		}

		m.failureHandlers = append(m.failureHandlers, make(chan worker.Worker, 1))
		m.failureHandlers[len(m.failureHandlers)-1] <- w
		m.handleFailures = true
		log.Println("created", c.Type, "worker as a failure handler")
	}
}

// ConfigStruct returns the struct used to configure the manager
func (m *Manager) ConfigStruct() interface{} {
	return config.DefaultAppConfig()
}

// Init initilize the manager
func (m *Manager) Init(i interface{}) error {
	conf, ok := i.(*config.AppConfig)
	if !ok {
		return config.WRONG_CONFIG_TYPE
	}
	m.numWorkers = &Counter{}
	m.currentConfig = conf
	m.jobChan = make(chan job.Job, 10)

	m.Stats = NewManagerStats(m)

	m.applyConfig(conf)

	// register handler
	m.statsServer.HandleFunc("/manager/stats", m.Stats.ReportStats)

	// set up all of the web servers
	go func() {
		log.Fatal(http.ListenAndServe(conf.StatsPort, m.statsServer))
	}()
	return nil
}

// NewManager create and return a referance to a manager
func NewManager() *Manager {
	m := &Manager{}
	m.KillChan = make(chan struct{})
	m.Providers = make(map[string]provider.Provider)
	m.allWorkers = make(map[uint64]worker.Worker)
	m.readyWorkers = make(map[string]chan worker.Worker)
	m.currentWorkers = make(map[string]int)
	m.statsServer = http.NewServeMux()
	return m
}

// remove a prefix from ever string in a slice
func trimStrings(s []string, pre string) []string {
	for i := 0; i < len(s); i++ {
		s[i] = strings.TrimPrefix(s[i], pre)
	}
	return s
}
