This sample demonstrates how to use Standalone Nexus Operations (executing Nexus operations directly from client code without wrapping them in a Workflow).
It shows both sync and async (workflow-backed) operations, and how to use the `ListNexusOperations` and `CountNexusOperations` APIs.

The starter and worker connect to two different namespaces (a "caller" namespace and a "target" namespace) — this mirrors how Nexus is typically used to cross namespace boundaries. The client is configured via the SDK's [environment configuration](https://docs.temporal.io/develop/environment-configuration) support, which reads `TEMPORAL_NAMESPACE`, `TEMPORAL_ADDRESS`, etc. from the environment (and optionally profiles from `temporal.toml`).

## Note: Standalone Nexus operations require a server version that supports this feature. Use the dev server build at https://github.com/temporalio/cli/releases/tag/v1.7.3-standalone-nexus-operations.

## Run locally against a dev server

1) Start the [Temporal dev server build that supports standalone Nexus operations](https://docs.temporal.io/standalone-nexus-operation#temporal-cli-support) with the required namespaces pre-created:

```bash
./temporal server start-dev \
  --namespace nexus-standalone-caller-namespace \
  --namespace nexus-standalone-target-namespace
```

2) Create a Nexus endpoint that routes to the target namespace and the worker's task queue:

```bash
./temporal operator nexus endpoint create \
  --name hello-service \
  --target-namespace nexus-standalone-target-namespace \
  --target-task-queue nexus-standalone-operations
```

3) In a new terminal, start the worker in the target namespace:

```bash
TEMPORAL_NAMESPACE=nexus-standalone-target-namespace \
  go run nexus-standalone-operations/worker/main.go
```

You should see a log line similar to:

```bash
2026/05/21 08:59:49 INFO  Started Worker Namespace nexus-standalone-target-namespace TaskQueue nexus-standalone-operations WorkerID 71172
```

4) In a third terminal, run the starter in the caller namespace:

```bash
TEMPORAL_NAMESPACE=nexus-standalone-caller-namespace \
  go run nexus-standalone-operations/starter/main.go
```

You should see something similar to:

```bash
2026/05/21 09:00:30 Started Echo operation OperationID nexus-standalone-echo-op
2026/05/21 09:00:30 Echo result: hello
2026/05/21 09:00:30 Started Hello operation OperationID nexus-standalone-hello-op
2026/05/21 09:00:30 Hello result: Hello Temporal 👋
2026/05/21 09:00:30 ListNexusOperations results:
2026/05/21 09:00:30     OperationID: nexus-standalone-hello-op, Operation: say-hello, Status: Completed
2026/05/21 09:00:30     OperationID: nexus-standalone-echo-op, Operation: echo, Status: Completed
2026/05/21 09:00:30 Total Nexus operations: 2
```

If you run the starter multiple times, additional entries will appear in the `ListNexusOperations` output and the `CountNexusOperations` total will grow.

## Run against Temporal Cloud

1) Create two namespaces in Temporal Cloud (for example `my-caller-namespace.<account>` and `my-target-namespace.<account>`) and generate an API key (or mTLS cert) that can access both.

2) Create a Nexus endpoint that routes to the target namespace and the worker's task queue. See the Temporal Cloud instructions at https://docs.temporal.io/nexus/registry#create-a-nexus-endpoint. Use:
- Endpoint name: `hello-service`
- Target namespace: `my-target-namespace.<account>`
- Target task queue: `nexus-standalone-operations`
- Allowed caller namespaces: include `my-caller-namespace.<account>` (endpoints reject callers that are not on this list)

3) Add two profiles to your [environment configuration file](https://docs.temporal.io/develop/environment-configuration), one per namespace. Using API keys:

```toml
[profile.target]
address = "<region>.<cloud>.api.temporal.io:7233"
namespace = "my-target-namespace.<account>"
api_key = "<your-api-key>"

[profile.caller]
address = "<region>.<cloud>.api.temporal.io:7233"
namespace = "my-caller-namespace.<account>"
api_key = "<your-api-key>"
```

For mTLS instead of API keys, set `tls.client_cert_path` and `tls.client_key_path` on each profile (see the [docs](https://docs.temporal.io/develop/environment-configuration) for the full schema).

4) Run the worker and starter in separate terminals, selecting the appropriate profile in each:

```bash
# terminal 1 (worker, target namespace)
export TEMPORAL_PROFILE="target"
go run nexus-standalone-operations/worker/main.go
```

```bash
# terminal 2 (starter, caller namespace)
export TEMPORAL_PROFILE="caller"
go run nexus-standalone-operations/starter/main.go
```
