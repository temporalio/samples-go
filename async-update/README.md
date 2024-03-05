### Async Update Sample

Here we show an example of a workflow representing a parallel job processor. The workflow accepts
jobs through update requests, allowing up to five parallel jobs, and uses the update validator to reject any
jobs over the limit. The workflow also demonstrates how to properly drain updates so all updates are processed before completing a workflow.

### Steps to run this sample:
1) Run a [Temporal service](https://github.com/temporalio/samples-go/tree/main/#how-to-use).
2) Run the following command to start the worker
```
go run async-update/worker/main.go
```
3) Run the following command to start the example
```
go run async-update/starter/main.go
```
