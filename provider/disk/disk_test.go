package disk

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/barracudanetworks/GoWorker/database"
	"github.com/barracudanetworks/GoWorker/job"
	"github.com/barracudanetworks/GoWorker/mock"
	"github.com/barracudanetworks/GoWorker/time_util"
)

var (
	testConfig = &DiskConfig{
		Name:   "test",
		Target: 10,
		DBName: "test.db",
		Bucket: "test",
	}
)

func TestFactory(t *testing.T) {
	d := DiskFactory()
	c := d.ConfigStruct()
	err := d.Init(c)
	if err != nil {
		t.Error(err)
	}
}

func TestRequestJob(t *testing.T) {
	d := diskHelper()
	err := addJobHelper(d)
	if err != nil {
		t.Error(err)
	}
	c := make(chan job.Job)
	go func() {
		err := d.RequestWork(1, c)
		if err != nil {
			t.Error(err)
		}
	}()
	j := <-c
	if j == nil {
		t.Fail()
	}
}

func diskHelper() *Disk {
	d := DiskFactory().(*Disk)
	d.Init(testConfig)
	return d
}

func addJobHelper(d *Disk) error {
	key := time_util.TimeToName(time.Now(), "test")
	b, err := json.Marshal(mock.NewMockJob())
	if err != nil {
		return err
	}

	return database.WriteJob(d.db, d.bucket, key, b)
}
