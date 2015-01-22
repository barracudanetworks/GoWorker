package worker

import (
	"errors"
	"log"
	"reflect"

	"strings"

	"github.com/barracudanetworks/GoWorker/config"
	"github.com/barracudanetworks/GoWorker/job"
)

// Factories holds all of the WorkerFacories that will be loaded for the manager to use
var (
	Factories        = make(map[string]WorkerFactory)
	WORKER_NOT_EXIST = errors.New("worker: worker type does not exist")
)

// LoadWorker loads a worker into the map of workers
func LoadWorker(w WorkerFactory) {
	v := reflect.TypeOf(w())

	name := strings.TrimPrefix(strings.ToLower(v.String()), "*")
	// set the worker factory by the name of the worker it produces
	Factories[name] = w
}

// Worker is used by a Manager to control the work being done by a sub-process
type Worker interface {
	// Work will be called by the manager to start the worker
	Work(j job.Job) *job.JobStats

	// Recycle get the worker ready to take on more work
	Recycle()

	// Kill send the kill signal to the worker
	Kill() error

	config.Configer
}

// WorkerFactory constructs and returns a new worker
type WorkerFactory func() Worker

// Create given a config create a new worker
func Create(t string, c config.Config) (Worker, error) {
	f, ok := Factories[t]
	if !ok {
		return nil, WORKER_NOT_EXIST
	}

	w := f()
	err := c.Apply(w)
	if err != nil {
		log.Println("Bad config for worker %s: %s", t, err)
		return nil, err
	}
	return w, nil
}
