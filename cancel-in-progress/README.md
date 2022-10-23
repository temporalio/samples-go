## Cancel-in-progress

This example demonstrates how to implement a workflow that ensures that only one run of a child workflow is executed at a time. Subsequent runs will cancel the runnign child workflow and re-run it with the latest options.
Those semantics are useful especially when implementing a CI pipeline. New commits during the execution of the workflow should short circuit the child workflow and only build the most recent commit.


Make sure the [Temporal Server is running locally](https://docs.temporal.io/docs/server/quick-install).

From the root of the project, start a Worker:

```bash
go run synchronous-build/worker/main.go
```

Start the Workflow Execution:

```bash
go run synchronous-build/starter/main.go
```
