This sample workflow demos context propagation through a workflow.

The workflow initializes the client with a context propagator which propagates
across the workflow - available to use in workflow and activities.

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
