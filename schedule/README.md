This sample demonstrates how to setup a schedule to run a workflow

For a configurable comparison of all Overlap Policies and Catchup Windows, see the [overlap sample](./overlap).

Steps to run this sample:
1) Run a [Temporal service](https://github.com/temporalio/samples-go/tree/main/#how-to-use).
2) Run 
```
go run schedule/worker/main.go 
```
to start worker for schedule workflow.
3) Run
```
go run schedule/starter/main.go
```
to start a schedule to run a workflow every second.
