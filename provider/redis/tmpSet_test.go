package redis

import (
	"encoding/json"
	"fmt"
	"log"
	"testing"

	"github.com/barracudanetworks/GoWorker/job"
)

var (
	test_prefix    = "test_prefix:"
	test_job_count = 0
	testJob        = func(r *Redis) *RedisJob {
		return r.createJob(testJobConfig())

	}
	testJobConfig = func() *job.JobConfig {
		test_job_count += 1
		return &job.JobConfig{
			Name:          "test",
			CaptureOutput: true,
			Params: json.RawMessage(`{
				"hello": "hello"
			}`),
			Type:    "cli",
			Retries: test_job_count,
		}
	}
)

func init() {
	log.SetFlags(log.Llongfile)

}

func TestPopAndLock(t *testing.T) {
	r, err := NewRedis("localhost:6379", 10, "job_list")
	if err != nil {
		t.Error(err)
	}
	addJobs(r, "job_list", 1)
	set, sErr := r.tmpSet.PopAndLock(r)
	if sErr != nil {
		t.Error(sErr)
	}

	if set == nil {
		t.Error("job not found")
	}
}

func TestGetAllOrphan(t *testing.T) {
	r, err := NewRedis("localhost:6379", 10, "job_list")
	if err != nil {
		t.Error(err)
	}

	addOrphanJob(r)
	_, err = r.tmpSet.GetOrphan(r, 1)
	if err != nil {
		t.Error(err)
	}

}

// addOrphanJob add an orphan job
func addOrphanJob(r *Redis) string {
	j := testJob(r)
	b, err := json.Marshal(j)
	if err != nil {
		log.Fatal(err)
	}
	r.Lock()
	_, err = r.conn.Do("set", fmt.Sprintf("tmp_job:value:%d", test_job_count), b)
	r.Unlock()
	if err != nil {
		log.Fatal(err)
	}
	return fmt.Sprintf("tmp_job:value:%d", test_job_count)
}
