package disk

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"hash"
	"log"
	"time"

	"github.com/barracudanetworks/GoWorker/config"
	"github.com/barracudanetworks/GoWorker/database"
	"github.com/barracudanetworks/GoWorker/job"
	"github.com/barracudanetworks/GoWorker/time_util"
	"github.com/barracudanetworks/GoWorker/worker"
	"github.com/boltdb/bolt"
)

const (
	DEFAULT_Disk = "my.db"
)

func init() {
	// load worker factory
	worker.LoadWorker(DiskFactory)
}

var (
	JOB_BUCKET = []byte("job_list")
)

// Disk a worker which writes jobs to a boltDisk
type Disk struct {
	db     *bolt.DB
	bucket []byte
	hasher hash.Hash
}

type DiskParams struct {
	ExicutionTime int64           `json:"exicution_time"`
	Job           json.RawMessage `json:"job"`
}

type DiskConfig struct {
	DB_Name string `json:"db_name" required:"true"`
	Bucket  string `json:"bucket" required:"true" description:"the bucket to insert jobs into"`
}

// Work write a job to disk using a bolt Disk
func (d *Disk) Work(j job.Job) *job.JobStats {
	stats := job.NewJobStats()

	// parse the config and write the job to disk
	err := d.writeJob(j)
	if err != nil {
		log.Println(err)
		stats.End(job.STATUS_FAILURE)
		return stats
	}

	stats.End(job.STATUS_SUCCESS)
	return stats
}

// getKey given a DiskParams object, create the key it corasponds to
func (d *Disk) getKey(p *DiskParams) []byte {
	return []byte(time_util.TimeToName(time.Unix(p.ExicutionTime, 0), fmt.Sprintf("%x", d.hasher.Sum(nil))))
}

// parseParams parse a DiskParams object from a raw jason message
func (d *Disk) parseParams(raw json.RawMessage) (*DiskParams, error) {
	params := &DiskParams{
		ExicutionTime: time.Now().Unix(),
	}
	err := json.Unmarshal(raw, params)
	return params, err
}

// Recycle get the worker ready for reuse
func (d *Disk) Recycle() {
	d.hasher.Reset()
}

// Kill this is a noop as the job can't be interupted
func (d *Disk) Kill() error {
	return nil

}

// ConfigStruct return the config structure for the DiskWorker
func (d *Disk) ConfigStruct() interface{} {
	return &DiskConfig{
		DB_Name: DEFAULT_Disk,
	}
}

// Init set the worker up for use
func (d *Disk) Init(i interface{}) error {
	conf, ok := i.(*DiskConfig)
	if !ok {
		return config.WRONG_CONFIG_TYPE
	}

	// attempt to open the database
	db, err := database.Open(conf.DB_Name)
	if err != nil {
		return err
	}
	d.hasher = sha1.New()

	d.db = db

	// create the bucket if it doesn't exist
	d.bucket = []byte(conf.Bucket)
	err = d.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(d.bucket)
		return err
	})

	return err
}

// writeJob write a job to the boltDisk
func (d *Disk) writeJob(j job.Job) error {
	conf := j.Config()

	// parse out the params object from the config
	params, err := d.parseParams(conf.Params)
	if err != nil {
		return err
	}

	key := d.getKey(params)

	err = database.WriteJob(d.db, d.bucket, key, params.Job)
	return err
}

// DiskFactory create and return a Disk worker
func DiskFactory() worker.Worker {
	return &Disk{}
}
