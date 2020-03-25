This sample workflow demos how to use query API to get the current state of running workflow.

`query_workflow.go` shows how to setup a custom workflow query handler

`query_workflow_test.go` shows how to unit-test query functionality

Steps to run this sample:
1) You need a Temporal service running. See details in README.md
2) Run the following command to start worker
```
go run query/worker/main.go
```
3) Run the following command to trigger a workflow execution. You should see workflowID and runID print out on screen.
```
go run query/starter/main.go
```
4) Run the following command to see current workflow state on the screen.
```
go run query/query/main.go
```
5) You could also specify the query type "__stack_trace" to dump the call stack for the workflow.
```
go run query/query/main.go -t __stack_trace
```
