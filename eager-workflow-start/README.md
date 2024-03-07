### Eager Workflow Start Sample

Eager Workflow Start (EWS) is an experimental latency optimization with the goal of reducing the time it takes to start a workflow.

When the starter and worker are collocated and aware of each other, they can interact while bypassing the server, saving a few time-intensive operations.

Here we modify the `helloworld` sample, ensuring that starter and worker run in the same process, and share the client. Also, the request eager mode flag is set in the start workflow options.

### Steps to run this sample:

1) Run a [Temporal service](https://github.com/temporalio/samples-go/tree/main/#how-to-use) with eager workflow start enabled by a dynamic config value.
```
temporal server start-dev --dynamic-config-value system.enableEagerWorkflowStart=true
```
2) Run the following command to start the combined worker and example
```
go run eager-workflow-start/main.go
```

