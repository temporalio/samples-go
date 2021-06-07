## Child Workflow Continue-As-New
<!-- @@@SNIPSTART samples-go-cw-cas-readme -->
This sample demonstrates that when a Child Workflow calls Continue-As-New it is not visible by a parent.
Parent Workflow Executions receive a notification that a Child Workflow Execution has completed only when the full execution has completed, failed, or timed out.

This feature is very useful when there is a need to process a large set of data.
The Child Execution can iterate over the data set, calling Continue-As-New periodically without polluting the parents' history.

Make sure the [Temporal Server is running locally](https://docs.temporal.io/docs/server/quick-install).

Start the Worker:

```bash
go run child-workflow-continue-as-new/worker/main.go
```

Start the Parent Workflow Execution:

```bash
go run child-workflow-continue-as-new/starter/main.go
```
<!-- @@@SNIPEND -->
