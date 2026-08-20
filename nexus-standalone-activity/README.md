## Nexus Operation Backed by a Standalone Activity

> [!WARNING]
> Standalone Nexus Operations and Standalone Activities are experimental and may be subject to backwards-incompatible changes. They require a Temporal server that implements and enables them via the dynamic configs shown below. Use the dev server build at https://github.com/temporalio/cli/releases/tag/v1.7.4-standalone-nexus-operations.

This sample shows how to implement a Nexus Operation whose backing execution is a **Standalone Activity**.

### Sample structure

| File                                       | Purpose                                                                                        |
|--------------------------------------------|------------------------------------------------------------------------------------------------|
| [`service/api.go`](./service/api.go)       | Nexus service definition shared by caller and handler                                          |
| [`handler/app.go`](./handler/app.go)       | The standalone Activity, and the operation built with `NewTemporalOperation` + `StartActivity` |
| [`worker/main.go`](./worker/main.go)       | Worker hosting the Nexus handler and the Activity                                              |
| [`starter/main.go`](./starter/main.go)     | Executes the Nexus Operation from client code                                                  |

The starter and worker connect to two different namespaces (a "caller" namespace and a "handler" namespace) — this mirrors how Nexus is typically used to cross namespace boundaries. The client is configured via the SDK's [environment configuration](https://docs.temporal.io/develop/environment-configuration) support, which reads `TEMPORAL_NAMESPACE`, `TEMPORAL_ADDRESS`, etc. from the environment (and optionally profiles from `temporal.toml`).

## Run locally against a dev server

1) Start the [Temporal dev server build that supports standalone Nexus Operations](https://docs.temporal.io/standalone-nexus-operation#temporal-cli-support) with the required namespaces pre-created and Activity callbacks enabled:

```bash
./temporal server start-dev \
  --dynamic-config-value activity.enableCallbacks=true \
  --namespace my-caller-namespace \
  --namespace my-handler-namespace
```

2) Create a Nexus endpoint that routes to the handler namespace and the worker's task queue:

```bash
./temporal operator nexus endpoint create \
  --name my-nexus-endpoint \
  --target-namespace my-handler-namespace \
  --target-task-queue nexus-handler-queue
```

3) In a second terminal, start the worker in the handler namespace:

```bash
TEMPORAL_NAMESPACE=my-handler-namespace \
  go run nexus-standalone-activity/worker/main.go
```

You should see a log line similar to:

```bash
2026/08/18 13:28:43 INFO  Started Worker Namespace my-handler-namespace TaskQueue nexus-handler-queue WorkerID 53608
```

4) In a third terminal, run the starter in the caller namespace:

```bash
TEMPORAL_NAMESPACE=my-caller-namespace \
  go run nexus-standalone-activity/starter/main.go
```

You should see something similar to:

```bash
2026/08/18 13:29:14 Started Greet operation OperationID greeting-afb1ff21-c842-40f7-ba85-28458d0150a6
2026/08/18 13:29:14 Hello, World!
```
