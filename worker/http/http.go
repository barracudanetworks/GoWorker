package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	std_http "net/http"
	"net/url"

	"github.com/barracudanetworks/GoWorker/job"
	"github.com/barracudanetworks/GoWorker/worker"
)

var (
	// default values
	DEFAULT_HEADERS = map[string]interface{}{}
	DEFAULT_PARAMS  = map[string]interface{}{}
)

func init() {
	// load http factory
	worker.LoadWorker(HttpFactory)
}

// Http a worker that makes std_http requests to external APIs
type Http struct {
	client       *std_http.Client
	currentJob   job.Job
	baseUrl      string
	doneChan     chan struct{}
	killChan     chan struct{}
	outputBuffer []byte
}

// HttpConfig provides config options for an http worker
type HttpConfig struct{}

// HttpParams provides params for an http worker
type HttpParams struct {
	Body      json.RawMessage        `json:"body"`
	Headers   map[string]interface{} `json:"headers"`
	Url       string                 `json:"url"`
	Method    string                 `json:"method"`
	UrlParams map[string]interface{} `jons:"url_params"`
}

func (h *Http) ConfigStruct() interface{} {
	return &HttpConfig{}
}

func (h *Http) Init(i interface{}) error {
	return nil
}

// Work perform an std_http request spesified by the given job
func (h *Http) Work(j job.Job) *job.JobStats {
	stats := job.NewJobStats()
	h.currentJob = j
	config := j.Config()
	params := &HttpParams{}
	err := json.Unmarshal(config.Params, params)
	if err != nil {
		log.Println(err, string(config.Params))
		stats.End(job.STATUS_FAILURE)
		return stats
	}

	r, err := h.genRequest(params)
	if err != nil {
		stats.End(job.STATUS_FAILURE)
		return stats
	}
	r.Header = generateHeader(params)

	// start new goroutine to make http call
	go func() {
		response, err := h.client.Do(r)
		if err != nil {
			log.Println(err)
			stats.End(job.STATUS_FAILURE)
		}
		if config.CaptureOutput {
			err = h.writeOutput(config.OutputWriter, response.Body)
			if err != nil {
				log.Println(err)
				stats.End(job.STATUS_FAILURE)
			}
		}
		select {
		case h.doneChan <- struct{}{}:
		default:
		}
	}()

	// wait for the http call to be finished, or for a kill signal to come in
	select {
	case <-h.killChan:
		stats.End(job.STATUS_FAILURE)
	case <-h.doneChan:
		if stats.Status() == job.STATUS_STARTED {
			stats.End(job.STATUS_SUCCESS)
		}
	}
	return stats
}

// writeOutput write out the response body to the a writer
func (h *Http) writeOutput(w io.Writer, r io.ReadCloser) error {
	var err error
	n := len(h.outputBuffer)
	for n == len(h.outputBuffer) {
		// read into buffer
		n, err = r.Read(h.outputBuffer)
		if err != nil {
			if err != io.EOF {
				r.Close()
				return err
			}
		}
		// write out buffer
		n, err = w.Write(h.outputBuffer[:n])
		if err != nil {
			r.Close()
			return err
		}
	}
	return nil
}

// generateRequest build and std_http.Request from a job config
func (h *Http) genRequest(p *HttpParams) (*std_http.Request, error) {
	buff := bytes.NewBuffer([]byte(p.Body))
	return std_http.NewRequest(p.Method, generateUrl(p), buff)
}

// generateHeader givin a JobConfig object, return std_http.Header
func generateHeader(p *HttpParams) std_http.Header {
	h := std_http.Header{}
	for k, v := range p.Headers {
		h.Add(k, fmt.Sprintf("%v", v))
	}
	return h
}

// generateUrl build a url string from a job config
func generateUrl(p *HttpParams) string {
	return p.Url + "?" + encodeValues(p.UrlParams)
}

// encodeValues create a urlencoded query string
func encodeValues(m map[string]interface{}) string {
	if len(m) == 0 {
		return ""
	}
	u := url.Values{}
	for k, v := range m {
		u.Add(k, fmt.Sprintf("%v", v))
	}
	return u.Encode()
}

// Kill kill the current process
func (h *Http) Kill() error {
	go func() { h.killChan <- struct{}{} }()
	return nil
}

// Recycle prepare this worker for it's next job
func (h *Http) Recycle() {
	h.currentJob = nil
}

// HttpFactory initilize and returns a new Worker of type std_http
func HttpFactory() worker.Worker {
	return NewHttp()
}

// Newstd_http initializes and returns a new std_http Worker
func NewHttp() *Http {
	return &Http{
		client:       &std_http.Client{},
		doneChan:     make(chan struct{}),
		killChan:     make(chan struct{}),
		outputBuffer: make([]byte, 1024),
	}
}
