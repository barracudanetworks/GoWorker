/*
Package GoWorker

A plugin based work management system implimented in pure go.
This project is still under heavy development. APIs are subject to change.

Installation

To install GoWorker, simply run
`go get github.com/barracudanetworks/GoWorker/...`

In order for this to work, you must have a valid GOPATH set
(read more about GOPATH [here](https://code.google.com/p/go-wiki/wiki/GOPATH))

Testing

To run all tests for the project simply navigate to the root of the project folder and run
`go test ./...`

To test an individual packages tests, navigate to the root of that package and run
`go test`

To run benchmarks contained in a package's tests run
`go test --bench=.`

More on testing in go can be found [here](http://www.golang-book.com/12/index.htm)

Usage

The main program for this project can be built by navigating to cmd/goworker/ and either running
`go run goworker.go` for testing purproses.

or

`go build` to compile the full binary for repeated use.

Composition

The work manager is compried of three main componants.
Provider
	Provides jobs to the manager from an external source.
Worker
	Takes a job, exicutes it, and returns its status and statistics about the job back to the manager.
Manager
	Manages the work pipline. Requests jobs from the providers and sends them to the workers.

Pipeline

There are three main stages in the worker pipeline.

1. The manager requests a job from the provider.
2. The Manager hands the job of to a worker that can handle the job.
3. Once the job is complete, the worker returns the job back to the manager, along with information about the result of the job. If the job was unsuccessful, and the job is configured to retry, step two and three will repeat until the job succeeds or the retries have been exhuasted. Once the retries have been exhausted, the manager hands the job off to the manager's failure handlers, if they have been configured.

So graphically...
__Provider->Manager->Worker->Manager__
where each arrow represents a transition of ownership of the job.

Plugins

Workers and providers are implemented as plugins. Because go compiles down to a static binary, you must recompile when you add new plugins.
To load your plugin into the binary, add a blank import to your package in the main go program inside of the cmd directory. Your package must have an init function that calls either `provider.LoadProvider(YourProviderFactory)` or `worker.LoadWorker(YourWorkerFactory)`. Once your worker or provider has been loaded, it will be accessable in the config file by its name, in all lowercase.

More on writing plugins http://godoc.org/github.com/barracudanetworks/GoWorker/worker http://godoc.org/github.com/barracudanetworks/GoWorker/provider
*/
package GoWorker
