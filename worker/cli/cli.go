package cli

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"

	"github.com/barracudanetworks/GoWorker/job"
	"github.com/barracudanetworks/GoWorker/worker"
)

func init() {
	worker.LoadWorker(CliFactory)
}

// Cli a worker that runs cli processes
type Cli struct {
	CurrentJob job.Job
	command    *exec.Cmd
	doneChan   chan struct{}
	killChan   chan struct{}
}

type CliConfig struct{}

type CliParams struct {
	Command string   `json:"command"`
	Params  []string `json:"params"`
}

// ConfigStruct return this workers config struct
func (c *Cli) ConfigStruct() interface{} {
	return &CliConfig{}
}

// Init set up the cli worker
func (c *Cli) Init(i interface{}) error {
	return nil
}

// Work run the cli process
func (c *Cli) Work(j job.Job) *job.JobStats {
	stats := job.NewJobStats()
	var err error
	c.CurrentJob = j
	config := j.Config()
	c.command, err = c.genCommand(config)
	if err != nil {
		log.Println(err)
		stats.End(job.STATUS_FAILURE)
		return stats
	}
	// if we need to capture the output, attach it to the stdout and stderr of the process
	if j.Config().CaptureOutput {
		c.command.Stdout = config.OutputWriter
		c.command.Stderr = config.OutputWriter
	}
	// start the proces
	err = c.command.Start()
	if err != nil {
		log.Println(err)
		stats.End(job.STATUS_FAILURE)
		return stats
	}

	// wait for the command to exit, and send back if there were any errors
	go func() {
		err := c.command.Wait()
		select {
		case c.doneChan <- struct{}{}:
			if err != nil {
				log.Println(err)
				stats.End(job.STATUS_FAILURE)
			}
		default:
		}
	}()

	// wait for a kill signal, or for the process to finish, and exit
	select {
	case <-c.killChan:
		err = c.command.Process.Kill()
		if err != nil {
			log.Println(err)
			stats.End(job.STATUS_FAILURE)
		}
		log.Println("Killed job ", j.Config())
		stats.End(job.STATUS_FAILURE)
	case <-c.doneChan:
		stats.End(job.STATUS_SUCCESS)
	}
	return stats
}

// Kill kill the current process
func (c *Cli) Kill() error {
	go func() { c.killChan <- struct{}{} }()
	return nil
}

// Recycle gets the Cli worker ready for more work
func (c *Cli) Recycle() {
	c.command = nil
}

// genCommand takes a job and returns a *exec.Cmd
func (c *Cli) genCommand(jc *job.JobConfig) (*exec.Cmd, error) {
	// parse out the params
	p := &CliParams{}
	err := json.Unmarshal(jc.Params, p)
	if err != nil {
		return nil, err
	}

	return exec.Command(p.Command, p.Params...), nil
}

// NewCli initializes and returns a new Cli Worker
func NewCli() *Cli {
	return &Cli{doneChan: make(chan struct{})}
}

// CliFactory initialzes and returns a new Worker
func CliFactory() worker.Worker {
	return NewCli()
}

// captureParams convert a map to an slice of strings
func captureParams(m map[string]interface{}) []string {
	s := make([]string, len(m))
	i := 0
	for _, v := range m {
		s[i] = fmt.Sprintf("%v", v)
		i += 1
	}
	return s
}
