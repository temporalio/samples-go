This sample shows how to use Temporal's Datadog interceptor to add tracing to your worker and clients.

### Setup

To run this sample make sure you have an active datadog agent reachable. This sample assume you have the agent running at `localhost:8126` if not adjust the sample accordingly.

https://docs.datadoghq.com/getting_started/agent/

#### Metrics

Starting with version 6.5.0 the Datadog agent is capable of scraping prometheus endpoints.

See more details here:
https://docs.datadoghq.com/integrations/guide/prometheus-host-collection/

Example `openmetrics.d/conf.yaml` to collect metrics from this sample and emit them to datadog under "myapp" namespace.

```
instances:
  - prometheus_url: http://localhost:9090/metrics
    namespace: "myapp"
    metrics:
      - temporal_datadog*
```

#### Logging

When using the Datadog interceptor all user created loggers will also include the trace and span ID. This makes
it possible to correlate logs with traces.

For documentation on how to configure your Datadog agent to upload logs see:
https://docs.datadoghq.com/logs/log_collection/go/

### Steps to run this sample:
1) Run a [Temporal service](https://github.com/temporalio/samples-go/tree/main/#how-to-use).
2) Run the following command to start the worker
```
go run datadog/worker/main.go
```
3) Run the following command to start the example
```
go run datadog/starter/main.go
