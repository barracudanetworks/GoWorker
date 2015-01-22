package http

import "github.com/barracudanetworks/GoWorker/job"
import std_http "net/http"

// HttpJob is a job created by the http provider
type HttpJob struct {
	config         *job.JobConfig
	provider       *Http
	responseWriter std_http.ResponseWriter
	request        *std_http.Request
	confirmChan    chan struct{}
}

// config return the JobConfig for this job
func (h *HttpJob) Config() *job.JobConfig {
	return h.config
}

// JobConfirmer return this job's provider
func (h *HttpJob) JobConfirmer() job.JobConfirmer {
	return h.provider
}
