This sample workflow demos context propagation through a workflow. Details about context propagation are
available [here](https://docs.temporal.io/docs/07_goclient/17_tracing).

The sample workflow initializes the client with a context propagator which propagates
specific information in the `context.Context` object across the workflow. The `context.Context` object is populated
with the information prior to calling `StartWorkflow`. The workflow demonstrates that the information is available
in the workflow and any activities executed.

Steps to run this sample:
1) You need a Temporal service running. See details README.md
2) Run the following command multiple times on different console window. This is to simulate running workers on multiple different machines.
```
go run ctxpropagation/worker/main.go
```
3) Run the following command to execute the context:
```
go run ctxpropagation/starter/main.go
```

You should see prints showing the context information available in the workflow
and activities.
