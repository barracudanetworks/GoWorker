package redis

import (
	"encoding/json"
	"log"
	"testing"
	"time"

	"github.com/barracudanetworks/GoWorker/job"
)

const (
	TEST_REDIS_HOST = "localhost"
	TEST_REDIS_PORT = ":6379"
)

var (
	testJson = []byte(`
		{"name":"test","capture_output":true,"params": {"hello": "hello"},"type": "cli","retries": 0}

	`)
	testRedisInfo = []byte(`
# Memory
used_memory:1013072
used_memory_human:989.33K
used_memory_rss:1851392
used_memory_peak:1101344
used_memory_peak_human:1.05M
used_memory_lua:41984
mem_fragmentation_ratio:1.83
mem_allocator:libc
		`)
	testRedisConfig = RedisConfig{
		Host:    TEST_REDIS_HOST,
		Port:    TEST_REDIS_PORT,
		JobList: "test_job_list",
	}
	testList   string
	testRedis  *Redis
	testConfig *job.JobConfig
)

func init() {
	testList = time.Now().String()
	r, err := NewRedis(TEST_REDIS_HOST+TEST_REDIS_PORT, 10, testList)
	if err != nil {
		log.Fatal(err)
	}
	testRedis = r

	testConfig = &job.JobConfig{
		Name:          "test",
		CaptureOutput: true,
		Params: json.RawMessage(`{
			"hello": "hello"
		}`),
		Type:    "cli",
		Retries: 0,
	}
}

func TestParseRedisInfoInt(t *testing.T) {
	i, err := parseRedisInfoInt(testRedisInfo, "used_memory")
	if err != nil {
		t.Error(err)
	}
	if i != 1013072 {
		t.Fail()
	}
}

func TestNewRedis(t *testing.T) {
	r, err := NewRedis("localhost:6379", 10, testList)
	if err != nil {
		t.Error(err)
	}
	if r == nil {
		t.Fail()
	}
	r.Close()
}

func TestNewRedisBadUrl(t *testing.T) {
	_, err := NewRedis("localhost:asdf", 10, testList)
	if err == nil {
		t.Fail()
	}
}

func TestPushJob(t *testing.T) {
	config, parseErr := job.ParseConfig(testJson)
	if parseErr != nil {
		t.Error(parseErr)
	}
	confirmConfig(config, t)
	j := testRedis.createJob(config)
	err := testRedis.pushJob(j, testList)
	if err != nil {
		t.Error(err)
	}
}

func TestProvider(t *testing.T) {
	p := testJob(testRedis).JobConfirmer()
	if p == nil {
		t.Error("Failed to get JobConfirmer")
	}
}

func TestParseJob(t *testing.T) {
	// upload test json
	testRedis.pushJob(testJob(testRedis), testList)
	j := testRedis.popJob()
	if j == nil {
		t.Error("popJob returned a bad job")
	}

	// confirm the job config was parsed correctly
	config := j.Config()
	confirmConfig(config, t)
}

func TestLenList(t *testing.T) {
	err := testRedis.pushJob(testJob(testRedis), testList+"TestLenList")
	if err != nil {
		t.Error(err)
	}
	n := testRedis.lenList(testList + "TestLenList")
	if n != 1 {
		t.Fail()
	}
}

func TestRequestWorkSingleJob(t *testing.T) {
	r, err := NewRedis("localhost:6379", 10, testList)
	if err != nil {
		t.Error(err)
	}
	r.pushJob(testJob(r), testList)
	jobChan := make(chan job.Job, 10)
	go r.RequestWork(1, jobChan)
	<-jobChan
	close(jobChan)
}

func TestRequestWork10Jobs(t *testing.T) {
	addJobs(testRedis, testList, 10)
	jobChan := make(chan job.Job)
	go testRedis.RequestWork(10, jobChan)
	// pull ten jobs from the server
	for i := 0; i < 10; i++ {
		<-jobChan
	}
	close(jobChan)
}

func TestCleanup(t *testing.T) {
	testRedis.conn.Do("flushall")
}

// helper function for adding and arbitrary  number of jobs to a list
func addJobs(r *Redis, listName string, numJobs int) {
	for i := 0; i < numJobs; i++ {
		r.pushJob(testJob(testRedis), listName)
	}
}

// helper function for confirming a job was parsed correctly
func confirmConfig(config *job.JobConfig, t *testing.T) {
	if !config.CaptureOutput {
		t.Error("Capture output is not correct")
	}
	if config.Name != "test" {
		t.Error("Name is not correct")
	}
	if config.Type != "cli" {
		t.Error("Type did not parse correctly")
	}
}
