package manager

import (
	"testing"
	"time"

	"github.com/eliothedeman/workhorse/mock"
	"github.com/eliothedeman/workhorse/provider"
)

func TestTrackerAdd(t *testing.T) {
	jt := NewAndStartTracker()

	j := mock.NewMockJob()
	w := &mock.MockWorker{}
	p := j.JobConfirmer().(provider.Provider)

	jt.Add(0, j, w, p)

	if jt.idToJob[0] != j {
		t.Fail()
	}

}

func TestTrackerRemove(t *testing.T) {
	jt := NewAndStartTracker()

	j := mock.NewMockJob()
	w := &mock.MockWorker{}
	p := j.JobConfirmer().(provider.Provider)

	jt.Add(0, j, w, p)

	if jt.idToJob[0] != j {
		t.Fail()
	}

	jt.Remove(j)

	//  give it time to do it's thing
	time.Sleep(1 * time.Millisecond)

	if _, inMap := jt.idToJob[0]; inMap {
		t.Fail()
	}
}

func TestTrackerStop(t *testing.T) {
	jt := NewAndStartTracker()

	j := mock.NewMockJob()
	w := &mock.MockWorker{}
	p := j.JobConfirmer().(provider.Provider)

	jt.Add(0, j, w, p)
}
