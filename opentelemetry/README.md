### Steps to run this sample:
1) Run a [Temporal service](https://github.com/temporalio/samples-go/tree/main/#how-to-use).

One way could be just to use the Temporal CLI.  

```bash
temporal server start-dev
```

2) Run the following command to start the worker
```bash
go run opentelemetry/worker/main.go
```
3) In another terminal, run the following command to run the workflow
```bash
go run opentelemetry/starter/main.go
```

The example outputs the traces in the stdout, both the worker and the starter.  

If all is needed is to see Workflows and Activities there's no need to set up instrumentation for the Temporal cluster.  

In order to send the traces to a real service you need to replace

```go
exp, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
if err != nil {
    log.Fatalln("failed to initialize stdouttrace exporter", err)
}
```
with  
```go
// Configure a new OTLP exporter using environment variables for sending data to Honeycomb over gRPC
clientOTel := otlptracegrpc.NewClient()
exp, err := otlptrace.New(ctx, clientOTel)
if err != nil {
    log.Fatalf("failed to initialize exporter: %e", err)
}
```

And provide the required additional parameters like the OTLP endpoint.  
For many services that would mean just to set the standard OTeL env vars like:

```
OTEL_SERVICE_NAME
OTEL_EXPORTER_OTLP_ENDPOINT
OTEL_EXPORTER_OTLP_HEADERS
```

As an example this is what is the rendered by Honeycomb.io.  

![Honeycomb.io](honeycomb_traces.png)
