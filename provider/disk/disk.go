package disk

import (
	"bytes"
	"encoding/json"
	"sync"
	"time"

	"github.com/barracudanetworks/GoWorker/config"
	"github.com/barracudanetworks/GoWorker/database"
	"github.com/barracudanetworks/GoWorker/job"
	"github.com/barracudanetworks/GoWorker/provider"
	"github.com/barracudanetworks/GoWorker/time_util"
	"github.com/boltdb/bolt"
)

func init() {
	provider.LoadProvider(DiskFactory)
}

// Disk a provider that uses a bolt db on disk as it's repository of jobs
type Disk struct {
	name      string
	db        *bolt.DB
	bucket    []byte
	tmpBucket []byte
	locks     *locker
	dbName    string
	target    float64
}

// DiskConfig the config struct used to set up the provider
type DiskConfig struct {
	Name   string  `json:"name" required:"true"`
	Target float64 `json:"target" required:"false"`
	DBName string  `json:"db_name" required:"false"`
	Bucket string  `json:"bucket" required:"true"`
}

// Locker holds locks for jobs
type locker struct {
	sync.Mutex
	l map[job.Job][]byte // map jobs to the keys in the tmp bucket
}

// NewLocker return a new locker
func NewLocker() *locker {
	return &locker{
		l: make(map[job.Job][]byte),
	}
}

// lock set a lock on the job
func (l *locker) lockJob(j job.Job, key []byte) {
	l.Lock()
	l.l[j] = key
	l.Unlock()
}

// unlock remove the lock on the job and return the key it corosponds to
func (l *locker) unlockJob(j job.Job) []byte {
	l.Lock()

	// get the key this job corosponds to
	key := l.l[j]

	// remove the lock on the job
	delete(l.l, j)
	l.Unlock()
	return key
}

// RequestWork collect work from the database
func (d *Disk) RequestWork(n int, jobChan chan job.Job) error {

	// get the time field for where we are scanning to
	t := []byte(time.Now().Format(time_util.TIME_FORMAT))

	// open up a read only view
	err := d.db.View(func(tx *bolt.Tx) error {
		// cycle though the jobs we have until we hit our limit
		for i := 0; i < n; i++ {
			// get a curser for our bucket
			c := tx.Bucket(d.bucket).Cursor()

			// hop to the first key
			c.First()

			// grab a key value pair
			k, v := c.Next()

			// if there is no key, we are done
			if k == nil {
				return nil
			}

			// get the time split from the hash value
			index := bytes.Index(k, []byte("#"))
			if index == -1 {
			}

			// if the time value is after now, end
			if bytes.Compare(k[0:index], t) > 0 {
				return nil
			}

			// parse and lock the job
			j, err := d.popAndLock(k, v)
			if err != nil {
				return err
			}

			// send the job back on the job chan
			jobChan <- j
		}
		return nil
	})

	return err
}

// ConfirmJob remove the job from the temporary list
func (d *Disk) ConfirmJob(j job.Job) error {
	return d.unlockJob(j)
}

// WaitTime return how long to wait before asking for more work
func (d *Disk) WaitTime(target float64) time.Duration {
	return 5 * time.Second
}

// Close gracefully shut down the provider
func (d *Disk) Close() error {
	err := database.Close(d.db)
	return err
}

// Target return the target jobs per second for this provider
func (d *Disk) Target() float64 {
	return d.target
}

// ConfigStruct
func (d *Disk) ConfigStruct() interface{} {
	return &DiskConfig{
		DBName: "my.db",
		Bucket: "job_list",
		Target: 20,
	}
}

// Init set up the disk provider
func (d *Disk) Init(i interface{}) error {

	// confirm the config struct
	conf, ok := i.(*DiskConfig)
	if !ok {
		return config.WRONG_CONFIG_TYPE
	}

	// set up the struct
	d.bucket = []byte(conf.Bucket)
	d.tmpBucket = []byte("tmp_" + conf.Bucket)
	d.name = conf.Name
	d.dbName = conf.DBName
	d.locks = NewLocker()

	// open the conection to the database
	db, err := database.Open(d.dbName)
	if err != nil {
		return err
	}

	// initilize bucket
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(d.bucket)
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists(d.tmpBucket)
		return err
	})

	if err != nil {
		return err
	}

	d.db = db

	return nil
}

// Name return the name of the provider
func (d *Disk) Name() string {
	return ""
}

// parseJob parse a job that points back to this provider
func (d *Disk) parseJob(b []byte) (job.Job, error) {

	// attempt to parse the buffer as a json blob into a JobConfig object
	conf := &job.JobConfig{}
	err := json.Unmarshal(b, conf)
	if err != nil {
		return nil, err
	}

	// create and return the job
	j := &DiskJob{
		provider: d,
		conf:     conf,
	}

	return j, nil
}

// popAndLock grab a job from the the database, and set a lock on it
func (d *Disk) popAndLock(k, v []byte) (job.Job, error) {

	// parse a job from the value
	j, err := d.parseJob(v)
	if err != nil {
		return nil, err
	}

	// set a lock for this job
	d.locks.lockJob(j, k)

	// put the job in the temporary bucket
	err = d.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(d.tmpBucket)
		return b.Put(k, v)
	})
	if err != nil {
		return j, err
	}

	// remove the job from the regular bucket
	err = d.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(d.bucket)
		return b.Delete(k)
	})
	if err != nil {
		return j, err
	}

	return j, nil
}

// unlockJob remove the lock from a job, and remove it from the temparary bucket
func (d *Disk) unlockJob(j job.Job) error {

	// remove the lock on the job
	k := d.locks.unlockJob(j)

	// remove the job from the temp bucket
	err := d.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(d.tmpBucket)
		return b.Delete(k)
	})

	return err
}

// DiskFactory create and return a disk provider
func DiskFactory() provider.Provider {
	return &Disk{}
}
