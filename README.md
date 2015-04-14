GoWorker
====
A plugin-based worker system implimented in Golang. APIs subject to change.

## Installation
1. Make sure the [GOPATH](https://code.google.com/p/go-wiki/wiki/GOPATH) environment variable is set.
2. Run `go get github.com/barracudanetworks/GoWorker/...`


## Testing
From the project root, run `go test ./...` to test all packages. To test a single package, run `go test` from within the package's directory.

To run benchmarks for a package, run `go test --bench=.`

[Read more about Golang testing](http://www.golang-book.com/12/index.htm)

## Usage
You may either run the program with `go run goworker.go` or build a binary with `go build`.

## Composition
GoWorker is comprised of three main components:

1. Provider: Provides jobs to the manager from an external source (e.g. Redis)
2. Worker: Executes a job and returns results and stats to the manager.
3. Manager: Manages the work pipline. Requests jobs from providers and routes them to the workers.

### Pipeline
__Provider -> Manager -> Worker -> Manager__ 

There are three main stages in the worker pipeline.

1. The Manager requests a job from a provider.
2. The Manager hands the job off to a worker.
3. The Worker returns the job back to the Manager after execution, with additional information including the result of the job. Jobs may optionally be set to retry in the case of failure until a maximum number of retries has been exhausted, or until the job succeeds. If all retries have been exhausted, the job is passed to the Manager's failure handlers, if any have been configured.

## Plugins
Workers and Providers are implemented as plugins. Because Go can compile down to a static binary, you must recompile GoWorker when you add new plugins.

To load your plugin into the binary, add a blank import to your package in the `main` package (`gowork.go` by default), located inside the `cmd` directory. Your package must have an `init()` function that calls either `provider.LoadProvider(YourProviderFactory)` or `worker.LoadWorker(YourWorkerFactory)`.

Once your worker or provider has been loaded, it will be accessible in the config file by its name, in all lowercase.

You may learn more by reading the GoDoc entries for [workers](http://godoc.org/github.com/barracudanetworks/GoWorker/worker) and [providers](http://godoc.org/github.com/barracudanetworks/GoWorker/provider).
