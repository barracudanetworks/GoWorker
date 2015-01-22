package http

import (
	"testing"

	"github.com/barracudanetworks/GoWorker/job"
	"github.com/barracudanetworks/GoWorker/mock"
)

var (
	testHttpWorker = &Http{}
	testJobConfig  = &job.JobConfig{
		Name:          "test",
		CaptureOutput: false,
		Type:          "http",
		Retries:       0,
	}
	basicGetServer       = mock.NewMockServer(mock.MockBasicGet)
	basicHeaderServer    = mock.NewMockServer(mock.MockBasicPrintHeaders)
	basicWaitServer      = mock.NewMockServer(mock.MockBasicWait)
	basicUrlValuesServer = mock.NewMockServer(mock.MockBasicUrlParams)
)

func init() {
	testHttpWorker = NewHttp()
}
func TestNewHttpWorker(t *testing.T) {
	h := NewHttp()
	if h == nil {
		t.Fail()
	}
}

func TestHttpFactory(t *testing.T) {
	h := HttpFactory()
	if h == nil {
		t.Fail()
	}
}

func TestHttpWork(t *testing.T) {
	s := testHttpWorker.Work(mock.NewBasicGetHttpJob(basicGetServer.Url()))
	if s.Status() != job.STATUS_SUCCESS {
		t.Fail()
	}
}

func TestHttpWorkWithHeaders(t *testing.T) {
	s := testHttpWorker.Work(mock.NewBasicWithHeadersJob(basicHeaderServer.Url()))
	if s.Status() != job.STATUS_SUCCESS {
		t.Fail()
	}
}

func TestHttpRecycle(t *testing.T) {
	testHttpWorker.Recycle()
}

func TestHttpKill(t *testing.T) {
	go testHttpWorker.Work(mock.NewBasicGetHttpJob(basicWaitServer.Url()))
	err := testHttpWorker.Kill()
	if err != nil {
		t.Error(err)
	}
}

func TestHttpUrlParams(t *testing.T) {
	s := testHttpWorker.Work(mock.NewBasicWithUrlParams(basicUrlValuesServer.Url()))
	if s.Status() != job.STATUS_SUCCESS {
		t.Fail()
	}
}
