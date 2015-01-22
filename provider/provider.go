package provider

import (
	"errors"
	"log"
	"reflect"
	"strings"
	"time"

	"github.com/barracudanetworks/GoWorker/config"
	"github.com/barracudanetworks/GoWorker/job"
)

var (
	Factories           = make(map[string]ProviderFactory)
	WRONG_CONFIG_TYPE   = errors.New("provider: wrong config type")
	PROVIDER_NOT_EXIST  = errors.New("provider: provider type does not exist")
	BAD_PROVIDER_CONFIG = errors.New("provider: bad config for provider")
)

// LoadProvider loads a provider factory
func LoadProvider(p ProviderFactory) {
	v := reflect.TypeOf(p())

	// get the name of the provider
	name := strings.TrimPrefix(strings.ToLower(v.String()), "*")
	Factories[name] = p
}

// Provider requests work from a datasource, and passes it to the manager
type Provider interface {
	// RequestWork tell the provider you are ready for new work
	RequestWork(int, chan job.Job) error
	// a provider must be able to confirm jobs as well
	job.JobConfirmer
	// WaitTime tells the manager how long to wait before checking for more work
	// target is given in terms of jobs per second
	WaitTime(target float64) time.Duration
	// Close tells the provider to gracefully close all of it's connections to the outside world
	Close() error
	Target() float64

	config.Configer

	Name() string
}

// ProviderFactory build and return a new provider
type ProviderFactory func() Provider

// Create given a config, create, initilize and return a new prover
func Create(name string, c config.Config) (Provider, error) {
	f, ok := Factories[name]
	if !ok {
		log.Println("could not find provider factory for %s", name)
		return nil, PROVIDER_NOT_EXIST
	}
	p := f()
	err := c.Apply(p)
	if err != nil {
		log.Fatalf("Bad config for provider %s: %s", name, err)
		return nil, BAD_PROVIDER_CONFIG
	}
	return p, nil
}
