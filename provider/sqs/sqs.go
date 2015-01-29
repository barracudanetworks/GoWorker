package sqs

import (
	"sync"
	"time"

	amazon_sqs "github.com/AdRoll/goamz/sqs"
	"github.com/eliothedeman/workhorse/config"
	"github.com/eliothedeman/workhorse/job"
	"github.com/eliothedeman/workhorse/provider"
)

const (
	DEFAULT_TIMEOUT = 30
)

func init() {
	provider.LoadProvider(SQSFactory)
}

// SQSConfig provides config information for the SQS provider
type SQSConfig struct {
	AccessKey string  `json:"access_key" required:"true"`
	SecretKey string  `json:"secret_key" required:"true"`
	Region    string  `json:"region" required:"true"`
	Queue     string  `json:"queue" required:"true"`
	Name      string  `json:"name" required:"false"`
	Timeout   int     `json:"timeout" timeout:"false"`
	Target    float64 `json:"target" required:"true"`
}

// SQS is a provider that allows for manual insertion of jobs
type SQS struct {
	accessKey, secretKey, region, name string
	timeout                            int
	conn                               *amazon_sqs.SQS
	queue                              *amazon_sqs.Queue
	jobs                               map[job.Job]*amazon_sqs.Message
	target                             float64
	sync.Mutex
}

func (s *SQS) ConfigStruct() interface{} {
	return &SQSConfig{
		Timeout: DEFAULT_TIMEOUT,
	}
}

// Init init a SQS provider
func (s *SQS) Init(i interface{}) error {
	s.Lock()
	defer s.Unlock()
	conf, ok := i.(*SQSConfig)
	if !ok {
		return config.WRONG_CONFIG_TYPE
	}

	// setup vars
	s.accessKey = conf.AccessKey
	s.secretKey = conf.SecretKey
	s.region = conf.Region
	s.name = conf.Name
	s.timeout = conf.Timeout
	s.target = conf.Target

	// set up the connection
	conn, err := amazon_sqs.NewFrom(s.accessKey, s.secretKey, s.region)
	if err != nil {
		return err
	}

	s.conn = conn

	// set up the queue
	var queue *amazon_sqs.Queue
	queue, err = s.conn.CreateQueueWithTimeout(conf.Queue, conf.Timeout)
	if err != nil {
		return err
	}

	s.queue = queue

	// SUCESS !!!
	return nil

}

// ConfirmJob
func (s *SQS) ConfirmJob(j job.Job) error {
	s.Lock()
	defer s.Unlock()
	// remove the message from sqs
	message, ok := s.jobs[j]
	if !ok {
		return job.JOB_NOT_FOUND
	}

	_, err := s.queue.DeleteMessage(message)
	delete(s.jobs, j)
	return err
}

// RequestWork register a new handler function. Then hold the provider open for ever.
func (s *SQS) RequestWork(n int, c chan job.Job) error {
	// get a back of messages from sqs
	messages, err := s.queue.ReceiveMessageWithVisibilityTimeout(n, s.timeout)
	if err != nil {
		return err
	}

	// for every message, parse out a job
	for i := range messages.Messages {
		body := []byte(messages.Messages[i].Body)
		conf, parseErr := job.ParseConfig(body)
		if parseErr != nil {
			return parseErr
		}

		j := job.NewBaseJob(s, conf, body)

		// send the job back on the given channel
		c <- j

		// start tracking the job locally
		s.Lock()
		s.jobs[j] = &messages.Messages[i]
	}
	return nil
}

// WaitTime
func (s *SQS) WaitTime(t float64) time.Duration {
	return 5 * time.Second
}

// Close
func (s *SQS) Close() error {
	return nil
}

// Target return 0.0 as this provider can not be limited
func (s *SQS) Target() float64 {
	return s.target
}

// Name return the name of the provider
func (s *SQS) Name() string {
	return "SQS_" + s.name
}

// SQSFactory factory for new SQS request
func SQSFactory() provider.Provider {
	return &SQS{}
}
