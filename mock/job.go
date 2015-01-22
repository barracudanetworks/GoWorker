package mock

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/barracudanetworks/GoWorker/job"
)

// MockJob is a job which exicutes `echo hello` as a cli command
type MockJob struct {
	confirmer job.JobConfirmer
	config    *job.JobConfig
}

// Config return a JobConfig for the test Job
func (m *MockJob) Config() *job.JobConfig {
	return m.config
}

// JobConfirmer return something that is able to confirm a job has been completed
func (m *MockJob) JobConfirmer() job.JobConfirmer {
	return m.confirmer
}

// NewMockJob initialize a new MockJob
func NewMockJob() *MockJob {
	return &MockJob{
		confirmer: &MockProvider{},
		config: &job.JobConfig{
			Retries:       0,
			Name:          "test",
			Params:        json.RawMessage(`{"command":"hello"}`),
			Type:          "cli",
			CaptureOutput: true,
			OutputWriter:  os.Stdout,
		},
	}
}

// NewMockHttpJob create a new job to be used in testing the worker/http package
func NewBasicGetHttpJob(url string) *MockJob {
	return &MockJob{
		confirmer: &MockProvider{},
		config: &job.JobConfig{
			Retries: 0,
			Name:    "http-test",
			Params: json.RawMessage(`{ 
				"url":        "` + url + `",
				"headers":    {},
				"url_params": {},
				"body":       "",
				"method":     "GET"
			}`),
			Type:          "http",
			CaptureOutput: false,
		},
	}
}

// NewBasicWithHeadersJob
func NewBasicWithHeadersJob(url string) *MockJob {
	return &MockJob{
		confirmer: &MockProvider{},
		config: &job.JobConfig{
			Retries: 0,
			Name:    "http-test",
			Params: json.RawMessage(`{ 
				"url":        "` + url + `",
				"headers":    {
					"my_header":"my_value"
					},
				"url_params": {},
				"body":       "",
				"method":     "GET"
			}`),
			Type:          "http",
			CaptureOutput: true,
			OutputWriter:  ioutil.Discard,
		},
	}
}

// NewBasicWithUrlParams
func NewBasicWithUrlParams(url string) *MockJob {
	return &MockJob{
		confirmer: &MockProvider{},
		config: &job.JobConfig{
			Retries: 0,
			Name:    "http-test",
			Params: json.RawMessage(`{ 
				"url":        "` + url + `",
				"headers":    {},
				"url_params": {
					"url":"value"
					},
				"body":       "",
				"method":     "GET"
			}`),
			Type:          "http",
			CaptureOutput: true,
			OutputWriter:  ioutil.Discard,
		},
	}
}

// NewFileJob
func NewFileJob(j job.Job) job.Job {
	b, err := json.Marshal(j.Config())
	if err != nil {
		log.Fatal(err)
	}
	return &MockJob{
		confirmer: &MockProvider{},
		config: &job.JobConfig{
			Retries: 0,
			Name:    "file-test",
			Params: json.RawMessage(`{
				"exicution_time": 123456789,
				"job": ` + string(b) + `
			}`),
			Type:          "file",
			CaptureOutput: false,
		},
	}
}

// NewDiskJob wrap a job in a disk job to be written to disk
func NewDiskJob(j job.Job) job.Job {
	b, err := json.Marshal(j.Config())
	if err != nil {
		log.Fatal(err)
	}
	return &MockJob{
		confirmer: &MockProvider{},
		config: &job.JobConfig{
			Retries: 0,
			Name:    "disk-test",
			Params: json.RawMessage(`{
				"bucket":  "test-bucket",
				"exicution_time": 123456,
				"job": ` + string(b) + `

			}`),
		},
	}
}
