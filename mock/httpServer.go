package mock

import (
	"fmt"
	"net/http"
	"net/http/httptest"
)

// MockHttpServer provides an http server for testing
type MockHttpServer struct {
	server *httptest.Server
}

// Url return the mock server's url
func (m *MockHttpServer) Url() string {
	return m.server.URL
}

// MockBasicGet is an http.handlerfunc that sends back hello world
func MockBasicGet(rw http.ResponseWriter, req *http.Request) {
	fmt.Fprintln(rw, "hello world")
}

// MockBasicPrintHeaders
func MockBasicPrintHeaders(rw http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(rw, "%v", req.Header)
}

// MockBasicWait this handler waits indeffinitly
func MockBasicWait(rw http.ResponseWriter, req *http.Request) {
	never := make(chan struct{})
	<-never
}

// MockBasicUrlParams
func MockBasicUrlParams(rw http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(rw, "%v", req.URL.Query())
}

// NewBasicGetServer creates a testing http server for basic get commands
func NewMockServer(f http.HandlerFunc) *MockHttpServer {
	m := &MockHttpServer{
		server: httptest.NewServer(http.HandlerFunc(f)),
	}
	return m
}
