This sample Workflow Definition demonstrates how to use multiple Temporal coroutines (instead of native goroutine) to process a
chunk of a large work item in parallel, and then merge the intermediate result to generate the final result.
In Temporal Workflow Definition, you should not use goroutines. Instead, you use corotinues via the `workflow.Go` method.

### Steps to run this sample:
1) You need a Temporal service running. See details in README.md
2) Run the following command to start the worker
```
go run splitmerge/worker/main.go
```
3) Run the following command to start the example
```
go run splitmerge/starter/main.go
```
