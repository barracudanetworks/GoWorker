package mock

import "github.com/barracudanetworks/GoWorker/job"

// MockWorker mocks out the worker interface for testing
type MockWorker struct {
}

// Work noop for testing
func (m *MockWorker) Work(j job.Job) *job.JobStats {
	stats := job.NewJobStats()
	stats.End(job.STATUS_SUCCESS)
	return stats
}

// Kill noop for testing
func (m *MockWorker) Kill() error {
	return nil
}
