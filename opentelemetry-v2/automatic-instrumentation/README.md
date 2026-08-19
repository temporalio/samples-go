## Automatic instrumentation

This sample enables Temporal SDK operation spans and SDK metrics on the Worker.

### Run the sample

1. Run a [Temporal Service](https://github.com/temporalio/samples-go/tree/main/#how-to-use).

2. Start Jaeger from the repository root:

```bash
docker compose -f opentelemetry-v2/docker-compose.yaml up -d
```

3. Start the Worker:

```bash
go run opentelemetry-v2/automatic-instrumentation/worker/main.go
```

4. In another terminal, start the Workflow:

```bash
temporal workflow execute \
  --task-queue automatic-instrumentation \
  --type Workflow \
  --input '"Temporal"'
```

5. Inspect the `temporal-otel-v2-automatic-worker` service in the
   [Jaeger UI](http://127.0.0.1:16686), and inspect the Temporal SDK metrics at
   [127.0.0.1:9090/metrics](http://127.0.0.1:9090/metrics).
