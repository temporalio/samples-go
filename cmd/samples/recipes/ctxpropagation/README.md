This sample workflow demos context propagation through a workflow. Details about context propagation are
available [here](https://cadenceworkflow.io/docs/03_goclient/16_tracing).

The sample workflow initializes the client with a context propagator which propagates
specific information in the `context.Context` object across the workflow. The `context.Context` object is populated
with the information prior to calling `StartWorkflow`. The workflow demonstrates that the information is available
in the workflow and any activities executed.

Steps to run this sample:
1) You need a cadence service running. See details in cmd/samples/README.md
2) Run the following command multiple times on different console window. This is to simulate running workers on multiple different machines.
```
./bin/ctxpropagation -m worker
```
3) Run the following command to execute the context .
```
./bin/ctxpropagation -m trigger
```

You should see prints showing the context information available in the workflow
and activities.
