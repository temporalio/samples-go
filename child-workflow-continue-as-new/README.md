## Child Workflow calling Continue As New

This sample demonstrates that a child workflow calling continue as new is not visible by a parent.
Parent receives notification about a child completion only when a child completes, fails or times out.

This is a useful feature when there is a need to process a large set of data. The child can iterate over the data set
calling continue as new periodically without polluting the parents' history.
 
### Steps to run this sample:
1) You need a Temporal service running. See details in README.md
2) Run the following command to start the worker
```
go run child-workflow-continue-as-new/worker/main.go
```
3) Run the following command to start the example
```
go run child-workflow-continue-as-new/starter/main.go
```
