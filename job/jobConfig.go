package job

import "io"
import "encoding/json"

// JobConfig configureation options for a job
type JobConfig struct {
	Name          string          `json:"name"`           // Name the name of the job
	CaptureOutput bool            `json:"capture_output"` // CaptureOutput if this is true, OutputWritter can not be nil!
	OutputWriter  io.Writer       `json:"-"`              // The writter that will be used to process the output (only used if CaptureOutput is true)
	Params        json.RawMessage `json:"params"`         // Params list of parameters to be given to the job at call time
	Type          string          `json:"type"`           // Type describes how this job can be run
	raw           []byte          // holds the raw job config to be used at a later time
	Retries       int             `json:"retries"` // Retries if this job fails, how many times should we retry
}

func (j *JobConfig) Raw() []byte {
	return j.raw
}

// ParseConfig takes a byte slice, atempts to parse it, and returns the new JobConfig
func ParseConfig(b []byte) (*JobConfig, error) {
	j := &JobConfig{}
	err := json.Unmarshal(b, j)
	j.raw = b
	return j, err
}
