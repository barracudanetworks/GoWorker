package manager

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/barracudanetworks/GoWorker/job"
	"github.com/barracudanetworks/GoWorker/provider"
)

// Counter provides methods for atomically counting things
type Counter struct {
	c uint64
}

// Inc add one to the counter
func (c *Counter) Inc() {
	atomic.AddUint64(&c.c, 1)
}

// Val returns the current count
func (c *Counter) Val() uint64 {
	return atomic.LoadUint64(&c.c)
}

// Set sets the counter to a given value
func (c *Counter) Set(val uint64) {
	atomic.StoreUint64(&c.c, val)
}

// DurationCounter tracks the adition of durations atomically
type DurationCounter struct {
	d time.Duration
	c *Counter
	sync.Mutex
}

// Add add two times together atomically
func (d *DurationCounter) Add(t time.Duration) {
	if d.c == nil {
		d.c = &Counter{}
	}
	d.Lock()
	d.d += t
	d.c.Inc()
	d.Unlock()
}

// Avg get the averge amount of time so far
func (d *DurationCounter) Avg() time.Duration {
	return d.Val() / time.Duration(d.c.Val())
}

// Val returns the current total
func (d *DurationCounter) Val() time.Duration {
	d.Lock()
	t := d.d
	d.Unlock()
	return t
}

func (d *DurationCounter) Reset() {
	d.Lock()
	d.d = 0
	d.c.Set(0)
	d.Unlock()
}

// Set set the counter to a given value
func (d *DurationCounter) Set(t time.Duration) {
	d.Lock()
	d.d = t
	d.Unlock()
}

// ManagerStats is a struct that holds statistics about a manager
type ManagerStats struct {
	lastTotalJobs            *Counter
	lastStatsCheck           time.Time
	lastStatsCheckByType     map[string]time.Time
	lastStatsCheckByProvider map[provider.Provider]time.Time
	lastTotalByType          map[string]*Counter
	lastTotalByProvider      map[provider.Provider]*Counter
	jobCountByType           map[string]*Counter
	jobCountByProvider       map[provider.Provider]*Counter
	jobDurationTotal         *DurationCounter
	jobDurationByType        map[string]*DurationCounter
	jobDurationByProvider    map[provider.Provider]*DurationCounter
	startTime                time.Time
	manager                  *Manager
}

// NewManagerStats initializes and returns a new instance of ManagerStats
func NewManagerStats(man *Manager) *ManagerStats {
	m := &ManagerStats{
		lastStatsCheck:           time.Now(),
		startTime:                time.Now(),
		lastTotalJobs:            &Counter{},
		jobCountByType:           make(map[string]*Counter),
		jobCountByProvider:       make(map[provider.Provider]*Counter),
		lastStatsCheckByType:     make(map[string]time.Time),
		lastStatsCheckByProvider: make(map[provider.Provider]time.Time),
		lastTotalByType:          make(map[string]*Counter),
		lastTotalByProvider:      make(map[provider.Provider]*Counter),
		jobDurationByType:        make(map[string]*DurationCounter),
		jobDurationByProvider:    make(map[provider.Provider]*DurationCounter),
		jobDurationTotal:         &DurationCounter{},
		manager:                  man,
	}
	return m
}

// IncrementJobs atomically add one to the total number of job
func (m *ManagerStats) IncrementJobs(j job.Job) {
	conf := j.Config()
	pro := j.JobConfirmer().(provider.Provider)
	// if we haven't seen this JobType before, create a zero value
	if _, inMap := m.jobCountByType[conf.Type]; !inMap {
		m.jobCountByType[conf.Type] = &Counter{}
	}

	// if we haven't seen this provider before, creat ea zero value
	if _, inMap := m.jobCountByProvider[pro]; !inMap {
		m.jobCountByProvider[pro] = &Counter{}
	}
	m.jobCountByType[conf.Type].Inc()
	m.jobCountByProvider[pro].Inc()
}

// JobsPerSecond return the job per second
func (m *ManagerStats) JobsPerSecond() float64 {
	job := m.TotalJobs() - m.lastTotalJobs.Val()
	elapsed := time.Now().Sub(m.lastStatsCheck)
	m.lastStatsCheck = time.Now()
	m.lastTotalJobs.Set(m.TotalJobs())
	return float64(job) / elapsed.Seconds()
}

// JobsPerSecondByType returns the job per second of a spesific type
func (m *ManagerStats) JobsPerSecondByType(t string) float64 {
	job := m.JobCountByType(t) - m.LastTotalByType(t)
	_, ok := m.lastStatsCheckByType[t]
	if !ok {
		m.lastStatsCheckByType[t] = time.Now()
	}
	elapsed := time.Now().Sub(m.lastStatsCheckByType[t])
	m.lastStatsCheckByType[t] = time.Now()
	m.lastTotalByType[t].Set(m.JobCountByType(t))
	return float64(job) / elapsed.Seconds()
}

// JobsPerSecondByProvider returns the job per second for a given Provider
func (m *ManagerStats) JobsPerSecondByProvider(p provider.Provider) float64 {
	job := m.JobCountByProvider(p) - m.LastTotalByProvider(p)
	_, ok := m.lastStatsCheckByProvider[p]
	if !ok {
		m.lastStatsCheckByProvider[p] = time.Now()
	}
	elapsed := time.Now().Sub(m.lastStatsCheckByProvider[p])
	m.lastStatsCheckByProvider[p] = time.Now()
	m.lastTotalByProvider[p].Set(m.JobCountByProvider(p))
	return float64(job) / elapsed.Seconds()
}

func (m *ManagerStats) LastTotalByType(t string) uint64 {
	_, ok := m.lastTotalByType[t]
	if !ok {
		m.lastTotalByType[t] = &Counter{}
	}

	return m.lastTotalByType[t].Val()
}

func (m *ManagerStats) LastTotalByProvider(p provider.Provider) uint64 {
	_, ok := m.lastTotalByProvider[p]
	if !ok {
		m.lastTotalByProvider[p] = &Counter{}
	}
	return m.lastTotalByProvider[p].Val()
}

// UpTime returns the uptime in seconds
func (m *ManagerStats) UpTime() time.Duration {
	return time.Now().Sub(m.startTime)
}

//TotalJobs returns the total number of job processed
func (m *ManagerStats) TotalJobs() uint64 {

	sum := uint64(0)
	for t, _ := range m.jobCountByType {
		sum += m.JobCountByType(t)
	}
	return sum
}

// TotalAverage returns the average number of job processed per second for the entire uptime
func (m *ManagerStats) TotalAverage() float64 {
	return float64(m.TotalJobs()) / m.UpTime().Seconds()
}

// AverageByType returns the average number of job processed per second for the entire uptime of this process, for a given type
func (m *ManagerStats) AverageByType(t string) float64 {
	return float64(m.JobCountByType(t)) / m.UpTime().Seconds()
}

func (m *ManagerStats) AverageByProvider(p provider.Provider) float64 {
	return float64(m.JobCountByProvider(p)) / m.UpTime().Seconds()
}

// TotalByType returns the total number of job for a given type
func (m *ManagerStats) JobCountByType(t string) uint64 {
	c, ok := m.jobCountByType[t]
	if ok {
		return c.Val()
	}
	return 0
}

func (m *ManagerStats) AverageDurationByType(t string) time.Duration {
	return m.JobDurationByType(t) / time.Duration(m.JobCountByType(t))
}

func (m *ManagerStats) AverageDurationByProvider(p provider.Provider) time.Duration {
	d := time.Duration(m.JobsPerSecondByProvider(p))
	if d == 0 {
		d = 1
	}
	return m.JobDurationByProvider(p) / d
}

func (m *ManagerStats) JobCountByProvider(p provider.Provider) uint64 {
	c, ok := m.jobCountByProvider[p]
	if ok {
		return c.Val()
	}
	return 0
}

func (m *ManagerStats) JobDurationByType(t string) time.Duration {
	c, ok := m.jobDurationByType[t]
	if ok {
		return c.Val()
	}
	return 0
}

func (m *ManagerStats) JobDurationByProvider(p provider.Provider) time.Duration {
	c, ok := m.jobDurationByProvider[p]
	if ok {
		return c.Val()
	}
	return 0
}

// collectJobsPerSecondByType
func (m *ManagerStats) collectJobsPerSecondByType() map[string]float64 {
	c := make(map[string]float64)
	for t, _ := range m.jobCountByType {
		c[t] = m.JobsPerSecondByType(t)
	}
	return c
}

// collectJobsPerSecondByProvider
func (m *ManagerStats) collectJobsPerSecondByProvider() map[string]float64 {
	c := make(map[string]float64)
	for p, _ := range m.jobCountByProvider {
		c[p.Name()] = m.JobsPerSecondByProvider(p)
	}
	return c
}

// collectJobTotalsByType
func (m *ManagerStats) collectJobTotalsByType() map[string]uint64 {
	c := make(map[string]uint64)
	for t, _ := range m.jobCountByType {
		c[t] = m.JobCountByType(t)
	}
	return c
}

// collectJobTotalsByProvider
func (m *ManagerStats) collectJobTotalsByProvider() map[string]uint64 {
	c := make(map[string]uint64)
	for p, _ := range m.jobCountByProvider {
		c[p.Name()] = m.JobCountByProvider(p)
	}
	return c
}

// collectTotalAveragesByType
func (m *ManagerStats) collectTotalAveragesByType() map[string]float64 {
	c := make(map[string]float64)
	for t, _ := range m.jobCountByType {
		c[t] = m.AverageByType(t)
	}
	return c
}

// collectTotalAveragesByProvider
func (m *ManagerStats) collectTotalAveragesByProvider() map[string]float64 {
	c := make(map[string]float64)
	for p, _ := range m.jobCountByProvider {
		c[p.Name()] = m.AverageByProvider(p)
	}
	return c
}

// collectTotalDurationsByType
func (m *ManagerStats) collectTotalDurationsByType() map[string]time.Duration {
	c := make(map[string]time.Duration)
	for t, _ := range m.jobDurationByType {
		c[t] = m.JobDurationByType(t)
	}
	return c
}

// collectTotalDurationsByProvider
func (m *ManagerStats) collectTotalDurationsByProvider() map[string]time.Duration {
	c := make(map[string]time.Duration)
	for p, _ := range m.jobDurationByProvider {
		c[p.Name()] = m.JobDurationByProvider(p)
	}
	return c
}

// collectAverageDurationsByType
func (m *ManagerStats) collectAverageDurationsByType() map[string]time.Duration {
	c := make(map[string]time.Duration)
	for t, _ := range m.jobDurationByType {
		c[t] = m.AverageDurationByType(t)
	}
	return c
}

// collectAverageDurationsByProvider
func (m *ManagerStats) collectAverageDurationsByProvider() map[string]time.Duration {
	c := make(map[string]time.Duration)
	for p, _ := range m.jobDurationByProvider {
		c[p.Name()] = m.AverageDurationByProvider(p)
	}
	return c
}

// consumeTime using a job.JobStats object. Modify the duration counters
func (m *ManagerStats) consumeTime(j job.Job, js *job.JobStats) {
	conf := j.Config()
	var d *DurationCounter
	p, ok := j.JobConfirmer().(provider.Provider)
	if !ok {
		log.Println(j.JobConfirmer(), "is not of correct type")
		return
	}

	// if we havent seen this provider before, create a duration tracker for it
	d, ok = m.jobDurationByProvider[p]
	if !ok {
		d = &DurationCounter{}
		m.jobDurationByProvider[p] = d
	}
	d.Add(js.Duration())

	// " type before, create a duration tracker for it
	d, ok = m.jobDurationByType[conf.Type]
	if !ok {
		d = &DurationCounter{}
		m.jobDurationByType[conf.Type] = d
	}
	d.Add(js.Duration())

	m.jobDurationTotal.Add(js.Duration())
}

// consumeStats takes a JobStats object, and applies it's stats to the ManagerStats
func (m *ManagerStats) consumeStats(j job.Job, js *job.JobStats) {
	m.IncrementJobs(j)
	m.consumeTime(j, js)
}

// collectChannelStats
func (m *ManagerStats) collectChannelStats() map[string]ChannelStats {
	chans := map[string]ChannelStats{
		"job_channel": ChannelStats{
			Capasity: cap(m.manager.jobChan),
			Queue:    len(m.manager.jobChan),
		},
	}
	for k, v := range m.manager.readyWorkers {
		chans["worker_"+string(k)] = ChannelStats{
			Capasity: cap(v),
			Queue:    cap(v) - len(v),
		}
	}
	return chans
}

// ChannelStats holds statistics about a channel
type ChannelStats struct {
	Capasity int `json:"capasity"`
	Queue    int `json:"queue"`
}

// collectStats creats a stats report fo the current state of the manager
func (m *ManagerStats) collectStats() ManagerStatsReport {
	msr := ManagerStatsReport{
		Uptime:                    uint64(m.UpTime() / time.Second),
		JobsPerSecond:             m.JobsPerSecond(),
		JobsPerSecondByType:       m.collectJobsPerSecondByType(),
		JobsPerSecondByProvider:   m.collectJobsPerSecondByProvider(),
		AverageDuration:           m.jobDurationTotal.Val() / (time.Duration(m.TotalJobs())),
		AverageDurationByType:     m.collectAverageDurationsByType(),
		AverageDurationByProvider: m.collectAverageDurationsByProvider(),
		TotalDurationByType:       m.collectTotalDurationsByType(),
		TotalDurationByProvider:   m.collectTotalDurationsByProvider(),
		TotalJobs:                 m.TotalJobs(),
		TotalJobsByType:           m.collectJobTotalsByType(),
		TotalJobsByProvider:       m.collectJobTotalsByProvider(),
		TotalAverage:              m.TotalAverage(),
		TotalAverageByType:        m.collectTotalAveragesByType(),
		TotalAverageByProvider:    m.collectTotalAveragesByProvider(),
		ChannelStats:              m.collectChannelStats(),
	}
	return msr
}

// ReportStats create a ManagerStatsReport for the current state of the manger
func (m *ManagerStats) ReportStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	msr := m.collectStats()
	b, err := json.Marshal(&msr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	_, err = w.Write(b)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}

// ManagerStatsReport holds statistical information about the mananger
type ManagerStatsReport struct {
	Uptime                    uint64                   `json:"uptime"`
	JobsPerSecond             float64                  `json:"job_per_second_cumulative"`
	JobsPerSecondByType       map[string]float64       `json:"job_per_second_by_type"`
	JobsPerSecondByProvider   map[string]float64       `json:"job_persecond_by_provider"`
	TotalJobs                 uint64                   `json:"total_job"`
	TotalJobsByType           map[string]uint64        `json:"total_job_by_type"`
	TotalJobsByProvider       map[string]uint64        `json:"total_job_by_provider"`
	TotalAverage              float64                  `json:"total_average"`
	TotalAverageByType        map[string]float64       `json:"total_average_by_type"`
	TotalAverageByProvider    map[string]float64       `json:"total_average_by_provider"`
	ChannelStats              map[string]ChannelStats  `json:"channel_stats"`
	AverageDurationByType     map[string]time.Duration `json:"average_duration_by_type"`
	AverageDurationByProvider map[string]time.Duration `json:"average_duration_by_provider"`
	AverageDuration           time.Duration            `json:"average_duration"`
	TotalDurationByType       map[string]time.Duration `json:"total_duration_by_type"`
	TotalDurationByProvider   map[string]time.Duration `json:"total_duration_by_provider"`
}
