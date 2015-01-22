package disk

import (
	"testing"

	"github.com/barracudanetworks/GoWorker/database"
	"github.com/barracudanetworks/GoWorker/job"
	"github.com/barracudanetworks/GoWorker/mock"
)

func TestDiskFactory(t *testing.T) {
	w := DiskFactory()

	// make sure we are getting the right type back
	if _, ok := w.(*Disk); !ok {
		t.Fail()
	}
}

func TestRecycle(t *testing.T) {
	w := DiskFactory()
	w.Init(w.ConfigStruct())
	w.Recycle()
}

func TestKill(t *testing.T) {
	w := DiskFactory()

	err := w.Kill()
	if err != nil {
		t.Error(err)
	}
}

func TestInit(t *testing.T) {
	w := DiskFactory()

	// get the empty config struct back
	conf := w.ConfigStruct().(*DiskConfig)
	conf.Bucket = "test_bucket"
	conf.DB_Name = "test.db"
	err := w.Init(conf)
	if err != nil {
		t.Error(err)
	}
}

func TestWork(t *testing.T) {
	w := DiskFactory().(*Disk)
	conf := w.ConfigStruct().(*DiskConfig)
	conf.DB_Name = "test.db"
	conf.Bucket = "test_bucket"
	err := w.Init(conf)
	if err != nil {
		t.Error(err)
	}

	j := mock.NewDiskJob(mock.NewMockJob())
	stats := w.Work(j)

	if stats.Status() != job.STATUS_SUCCESS {
		t.Fail()
	}
	var p *DiskParams
	p, err = w.parseParams(j.Config().Params)
	if err != nil {
		t.Error(err)
	}
	// read the job
	jc, jErr := database.ReadJob(w.db, w.bucket, w.getKey(p))
	if jErr != nil {
		t.Error(err)
	}

	if string(jc.Params) != string(mock.NewMockJob().Config().Params) {
		t.Fail()
	}
}
