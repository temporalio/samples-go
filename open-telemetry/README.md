```
docker run --name jaeger \
  -e COLLECTOR_OTLP_ENABLED=true \
  -p 16686:16686 \
  -p 4317:4317 \
  -p 4318:4318 \
  jaegertracing/all-in-one:latest
```

## Open Telemetry Workflow
Make sure the [Temporal Server is running locally](https://docs.temporal.io/application-development/foundations#run-a-development-cluster).

From the root of the project, start a Worker:

```bash
go run open-telemetry/worker/main.go
```

Start the Workflow Execution:

```bash
go run open-telemetry/starter/main.go
```
