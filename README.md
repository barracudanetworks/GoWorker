goworker
====
GoWorker is still under heavy development. APIs are subject to change.

## Installation
To install GoWorker, simply run 
`go get github.com/barracudanetworks/GoWorker/...`

#### Hint
in order for this to work, you must have a valid GOPATH set
(read more about GOPATH [here](https://code.google.com/p/go-wiki/wiki/GOPATH))

## Testing
To run all tests for the project simply navigate to the root of the project folder and run
`go test ./...`

To test an individual packages tests, navigate to the root of that package and run
`go test`

To run benchmarks contained in a package's tests run
`go test --bench=.`

More on testing in go can be found [here](http://www.golang-book.com/12/index.htm)

## Usage
The main program for this project can be built by navigating to cmd/goworker/ and either running 
`go run goworker.go` for testing purproses.

or

`go build` to compile the full binary for repeated use. 
