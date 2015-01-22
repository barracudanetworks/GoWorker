package manager

import (
	"testing"

	"github.com/barracudanetworks/GoWorker/mock"
)

func TestIncrementStats(t *testing.T) {
	m := NewManagerStats(TEST_MANAGER)
	j := mock.NewMockJob()
	m.IncrementJobs(j)

	if m.TotalJobs() != 1 {
		t.Error("Increment of jobs failed")
	}

	if m.JobCountByType("cli") != 1 {
		t.Error("Failed to increment of spesific job type")
	}

	if m.JobCountByType("http") == 1 {
		t.Error("Wrong job type incremented")
	}
}

func TestJobsPerSecond(t *testing.T) {
	m := NewManagerStats(TEST_MANAGER)
	j := mock.NewMockJob()
	m.IncrementJobs(j)

	if m.TotalAverage() < 0 {
		t.Fail()
	}

	if m.AverageByType("cli") < 0 {
		t.Fail()
	}

	if m.JobsPerSecond() < 0 {
		t.Fail()
	}

	if m.JobsPerSecondByType("cli") < 0 {
		t.Fail()
	}
}
