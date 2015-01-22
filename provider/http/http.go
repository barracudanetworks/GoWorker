/*
Package http provider allows for jobs to be manualy inserted into the manager though an http api.
This provider can be enabled though the config file, just as any other provider. However, the manager will not pull this provider for jobs.
*/
package http

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	std_http "net/http"
	"time"

	"github.com/barracudanetworks/GoWorker/job"
	"github.com/barracudanetworks/GoWorker/provider"
)

const (
	DEFAULT_ENDPOINT = "/job/add"
)

var (
	STATUS_TO_HTTP_CODE = map[job.Status]uint{
		job.STATUS_SUCCESS: std_http.StatusOK,
		job.STATUS_FAILURE: std_http.StatusInternalServerError,
	}

	NOT_HTTP_JOB = errors.New("This job is not of the correct type. Expecting type *HttpJob")
)

func init() {
	// load http provider
	if provider.Factories == nil {
		log.Println("unable to load provider")
	} else {
		provider.Factories["http"] = HttpFactory
	}
}

// HttpConfig provides config information for the Http provider
type HttpConfig struct {
	ListOn   string `json:"listen_on"`
	Endpoint string `json:"endpoint"`
}

// Http is a provider that allows for manual insertion of jobs
type Http struct {
	jobChan   chan job.Job
	endPoint  string
	listen_on string
	server    *std_http.ServeMux
}

func (h *Http) ConfigStruct() interface{} {
	return &HttpConfig{}
}

// Init init a http provider and start it's webserver
func (h *Http) Init(i interface{}) error {
	conf, ok := i.(*HttpConfig)
	if !ok {
		return provider.WRONG_CONFIG_TYPE
	}
	h.endPoint = conf.Endpoint
	h.server = std_http.NewServeMux()
	h.server.HandleFunc(conf.Endpoint, h.addNewJob)
	h.listen_on = conf.ListOn

	// handle the function
	go func() {
		log.Fatal(std_http.ListenAndServe(conf.ListOn, h.server))
	}()
	return nil
}

// ConfirmJob
func (h *Http) ConfirmJob(j job.Job) error {
	// assume that all jobs coming into this confirmer are of type HttpJob
	hj, ok := j.(*HttpJob)
	if !ok {
		return NOT_HTTP_JOB
	}
	// make sure the http response is taken care of
	<-hj.confirmChan
	return nil
}

// RequestWork register a new handler function. Then hold the provider open for ever.
func (h *Http) RequestWork(n int, c chan job.Job) error {
	h.jobChan = c

	<-make(chan struct{})
	return nil
}

// WaitTime
func (h *Http) WaitTime(t float64) time.Duration {
	return 0 * time.Second
}

// Close
func (h *Http) Close() error {
	return nil
}

// Target return 0.0 as this provider can not be limited
func (h *Http) Target() float64 {
	return 0
}

// Name return the name of the provider
func (h *Http) Name() string {
	return "http_" + h.listen_on + h.endPoint
}

// addNewJob an http.Handlerfunc that parses a job from the body of a request, and sends it on the jobChan
func (h *Http) addNewJob(rw std_http.ResponseWriter, req *std_http.Request) {
	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		fmt.Fprint(rw, err)
	}

	var jc *job.JobConfig
	jc, err = job.ParseConfig(b)
	if err != nil {
		log.Println(err, b)
		fmt.Fprint(rw, err)
		return
	}
	// make the job able to write back to the requester
	jc.OutputWriter = rw

	// create a new HttpJob and send it to the manager
	j := &HttpJob{
		config:         jc,
		provider:       h,
		responseWriter: rw,
		request:        req,
		confirmChan:    make(chan struct{}),
	}
	h.jobChan <- j

	// hold this request open until the response is ready to be written
	j.confirmChan <- struct{}{}
}

// HttpFactory factory for new http request
func HttpFactory() provider.Provider {
	return &Http{}
}
