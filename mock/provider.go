package mock

import (
	"time"

	"github.com/barracudanetworks/GoWorker/job"
)

// MockProvider is a testing job provider
type MockProvider struct {
}

// RequestWork make fake request for work, launch provideWork
func (m *MockProvider) RequestWork(numJobs int, j chan job.Job) error {
	go m.provideWork(numJobs, j)
	return nil
}

// provideWork provide fake work
func (m *MockProvider) provideWork(numJobs int, j chan job.Job) {
	for i := 0; i < numJobs; i++ {
		j <- NewMockJob()
	}
}

// ConfirmJob confirm that the job has been completed
func (m *MockProvider) ConfirmJob(j job.Job) error {
	return nil
}

// WaitTime tell the manager how long to wait for work
func (m *MockProvider) WaitTime(target float64) time.Duration {
	return 5 * time.Second
}

// Close close the mock provider
func (m *MockProvider) Close() error {
	return nil
}

// Target
func (m *MockProvider) Target() float64 {
	return 0
}

// ConfigStruct return the mock provider config struct
func (m *MockProvider) ConfigStruct() interface{} {
	return struct{}{}
}

// Init initialize the MockProvider
func (m *MockProvider) Init(i interface{}) error {
	return nil
}

// Name
func (m *MockProvider) Name() string {
	return "mock"
}
