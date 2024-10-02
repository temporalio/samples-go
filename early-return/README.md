### Early-Return Sample

This sample demonstrates an early-return from a workflow.

By utilizing Update-with-Start, a client can start a new workflow and synchronously receive 
a response mid-workflow, while the workflow continues to run to completion.

### Steps to run this sample:
1) Run a [Temporal service](https://github.com/temporalio/samples-go/tree/main/#how-to-use).
2) Run the following command to start the worker
```
go run early-return/worker/main.go
```
3) Run the following command to start the example
```
go run early-return/starter/main.go
```
