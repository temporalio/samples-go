This sample Workflow demos context propagation through a Workflow. Details about context propagation are
available [here](https://docs.temporal.io/docs/go/tracing).

The sample Workflow initializes the client with a context propagator which propagates
specific information in the `context.Context` object across the Workflow. The `context.Context` object is populated
with the information prior to calling `StartWorkflow`. The Workflow demonstrates that the information is available
in the Workflow and any activities executed.

Also, this sample initializes a Jaeger global tracer and pass it to the client. The sample will work without
actual Jaeger instance -- just report every tracer call to the log. To see traces in Jaeger run it with follow command:
```
$ docker run --publish 6831:6831 --publish 16686:16686 jaegertracing/all-in-one:latest
```

Steps to run this sample:
1) You need a Temporal service running. See details README.md.
2) Run
```
go run ctxpropagation/worker/main.go
```
to start worker for `ctxpropagation` Workflow.

3) Run:
```
go run ctxpropagation/starter/main.go
```
to start Workflow.

You should see prints showing the context information available in the workflow and activities.
