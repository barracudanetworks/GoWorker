package database

import (
	"sync"

	"github.com/barracudanetworks/GoWorker/job"
	"github.com/boltdb/bolt"
)

var (
	dbContainer = &container{
		dbs: make(map[string]*holder),
	}
)

func Open(filename string) (*bolt.DB, error) {
	dbContainer.Lock()
	defer dbContainer.Unlock()

	// if we already have this file open, return the reference to the db
	if db, inMap := dbContainer.dbs[filename]; inMap {
		db.users += 1
		return db.db, nil
	}

	// open a new database
	db, err := bolt.Open(filename, 0660, nil)
	if err != nil {
		return nil, err
	}

	// create a holder object and initilize the "users" field to 1 because we now have one user of the connection
	h := &holder{
		db:    db,
		users: 1,
	}
	dbContainer.dbs[filename] = h
	return db, nil
}

func Close(db *bolt.DB) error {
	dbContainer.Lock()
	defer dbContainer.Unlock()
	if db, inMap := dbContainer.dbs[db.Path()]; inMap {
		db.users -= 1

		// if there are no more users for this db, remove it from the map completely

		if db.users < 1 {
			delete(dbContainer.dbs, db.db.Path())
			err := db.db.Close()
			return err
		}
		return nil
	}
	return nil
}

// ReadJob read a JobConfig object from a bolt db
func ReadJob(db *bolt.DB, bucket, key []byte) (conf *job.JobConfig, err error) {
	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucket)
		buff := b.Get([]byte(key))
		conf, err = job.ParseConfig(buff)

		return err
	})
	return
}

// WriteJob write a job to disk
func WriteJob(db *bolt.DB, bucket, key, data []byte) error {
	err := db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists(bucket)
		if err != nil {
			return err
		}
		if b == nil {
			b = tx.Bucket(bucket)

		}
		return b.Put(key, data)
	})
	return err
}

type container struct {
	dbs map[string]*holder
	sync.Mutex
}

type holder struct {
	db    *bolt.DB
	users int
}
