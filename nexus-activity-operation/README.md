# nexus-activity-operation

This sample shows how to author a Nexus operation **backed by an activity** using the generic
`temporalnexus.NewTemporalOperation` constructor, and how to invoke it directly using the
standalone `client.NexusClient` (no caller workflow required).

Activity-backed Nexus operations let the handler execute a single, long running, side-effecting call —
an API request, a database write, a compute step — without paying the cost of a workflow
execution. Temporal's retry, timeout, and cancellation semantics still apply via standard
activity options.

`NewTemporalOperation` is the unified constructor for Temporal-backed Nexus operations. The
`Start` callback receives a `NexusClient` scoped to the invocation and returns a
`TemporalOperationResult` produced by `temporalnexus.StartActivity` (or `StartUntypedActivity`
when the activity doesn't follow the single-input / single-output signature). The same
constructor can back operations with workflows (`StartWorkflow` / `StartUntypedWorkflow`) and
expose `CancelWorkflowRun` / `CancelActivityExecution` hooks for customizing cancel behavior
per token type.

The starter uses `client.NewNexusClient` + `ExecuteOperation` — the standalone Nexus API for
invoking operations from non-workflow code. For invocations from inside a workflow, use
`workflow.NewNexusClient` (see the [`nexus`](../nexus) sample).

This sample defines a single operation:

- `say-hello` — an asynchronous operation that schedules `HelloHandlerActivity` via
  `temporalnexus.StartActivity` and returns an async result.

### Sample directory structure

- [service](./service) - shared service definition
- [handler](./handler) - operations and worker, defined with `NewTemporalOperation`
- [starter](./starter) - standalone Nexus client that invokes the operation
- [options](./options) - command line argument parsing utility

## Getting started locally

### Get `temporal` CLI to enable local development

Follow the [docs site](https://learn.temporal.io/getting_started/go/dev_environment/#set-up-a-local-temporal-service-for-development-with-temporal-cli)
to install Temporal CLI (v1.3.0 or later recommended).

### Spin up environment

#### Start temporal server

Activity-backed Nexus operations are currently gated behind a server dynamic config flags. For this sample please start the dev server with these three enabled:

```
temporal server start-dev \
  --dynamic-config-value frontend.activityAPIsEnabled=true \
  --dynamic-config-value nexusoperation.enableStandalone=true \
  --dynamic-config-value activity.enableCallbacks=true
```

- `frontend.activityAPIsEnabled` exposes the Standalone Activity client APIs that the
  handler uses to schedule the backing activity.
- `nexusoperation.enableStandalone` enables Standalone Nexus operations.
- `activity.enableCallbacks` enables callbacks on activities so the activity completion
  can be delivered back through the Nexus operation token.

In a separate terminal:

#### Create caller and target namespaces

```
temporal operator namespace create --namespace my-target-namespace
temporal operator namespace create --namespace my-caller-namespace
```

#### Create Nexus endpoint

> NOTE: this must be run in the `nexus-activity-operation` sample directory.

```
temporal operator nexus endpoint create \
  --name my-nexus-endpoint-name \
  --target-namespace my-target-namespace \
  --target-task-queue my-handler-task-queue \
  --description-file ./service/description.md
```

### Run the handler worker

```
cd handler
go run ./worker \
    -target-host localhost:7233 \
    -namespace my-target-namespace
```

### Invoke the operation

```
cd starter
go run . \
    -target-host localhost:7233 \
    -namespace my-caller-namespace
```

### Output

```
Hello result: ¡Hola! Nexus 👋
```
