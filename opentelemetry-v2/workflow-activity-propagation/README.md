## Workflow-to-Activity propagation

This sample propagates a replay-safe custom Workflow span to a custom Activity
span.

### Run the sample

1. Run a [Temporal Service](https://github.com/temporalio/samples-go/tree/main/#how-to-use).

2. Start Jaeger from the repository root:

```bash
docker compose -f opentelemetry-v2/docker-compose.yaml up -d
```

3. Start the Worker:

```bash
go run opentelemetry-v2/workflow-activity-propagation/worker/main.go
```

4. In another terminal, start the Workflow:

```bash
temporal workflow execute \
  --task-queue workflow-activity-propagation \
  --type Workflow \
  --input '"Temporal"'
```

5. Inspect the
   `temporal-otel-v2-custom-workflow-activity-propagation-worker` service in the
   [Jaeger UI](http://127.0.0.1:16686).
