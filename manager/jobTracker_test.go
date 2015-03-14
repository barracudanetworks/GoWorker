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
	//  give it time to do it's thing
	time.Sleep(1 * time.Millisecond)

	if jt.jobToPack[j].job != j {
		t.Fail()
	}

}

func TestByAge(t *testing.T) {
	jt := NewAndStartTracker()

	j := mock.NewMockJob()
	w := &mock.MockWorker{}
	p := j.JobConfirmer().(provider.Provider)

	jt.Add(0, j, w, p)

	l := jt.ByAge(time.Now())

	if len(l) != 1 {
		t.Fail()
	}
}

func TestTrackerRemove(t *testing.T) {
	jt := NewAndStartTracker()

	j := mock.NewMockJob()
	w := &mock.MockWorker{}
	p := j.JobConfirmer().(provider.Provider)

	jt.Add(0, j, w, p)
	//  give it time to do it's thing
	time.Sleep(1 * time.Millisecond)

	if jt.jobToPack[j].job != j {
		t.Fail()
	}

	jt.Remove(j)

	//  give it time to do it's thing
	time.Sleep(1 * time.Millisecond)

	if _, inMap := jt.jobToPack[j]; inMap {
		t.Fail()
	}
}

func TestTrackerStop(t *testing.T) {
	jt := NewAndStartTracker()

	j := mock.NewMockJob()
	w := &mock.MockWorker{}
	p := j.JobConfirmer().(provider.Provider)
	//  give it time to do it's thing
	time.Sleep(1 * time.Millisecond)

	jt.Add(0, j, w, p)
}
