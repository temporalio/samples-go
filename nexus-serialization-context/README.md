# Nexus serialization context

This sample uses `converter.NexusSerializationContext` to select an encryption key for
each Nexus endpoint. The caller invokes synchronous and workflow-backed asynchronous
operations through four endpoints. The `compressed-encrypted-worker` compresses and then
encrypts Nexus payloads. The `encrypted-worker` encrypts without compression. Each handler
selects between its own key 1 and key 2 using the endpoint in the Nexus context.
The caller and each handler run in separate namespaces.

| Handler worker | Endpoint | Key ID | Nexus payload encoding |
|---|---|---|---|
| Compressed/encrypted | `nexus-serialization-compressed-encrypted-1` | `compressed-encrypted-1` | zlib, then AES-GCM |
| Compressed/encrypted | `nexus-serialization-compressed-encrypted-2` | `compressed-encrypted-2` | zlib, then AES-GCM |
| Encrypted | `nexus-serialization-encrypted-1` | `encrypted-1` | AES-GCM |
| Encrypted | `nexus-serialization-encrypted-2` | `encrypted-2` | AES-GCM |

These four endpoint keys are distinct 128-bit AES keys. Compression runs before encryption;
decoding decrypts before decompressing. AES-GCM authenticates the ciphertext.
The AES-GCM codec is implemented in this sample; compression uses the SDK's zlib codec.
Nexus supplies the endpoint context used to select them. The context also identifies the
service and operation, which a codec can use to select keys instead of, or along with,
the endpoint.

`nexusCodec` implements `converter.PayloadCodecWithSerializationContext`. Each worker
configures it with an endpoint-to-codec map and a default codec for ordinary workflow
payloads. The caller uses a separate key for its own input and
final result, keeping those payloads independent of the keys shared with Nexus handlers.
Each handler reuses its key 1 for backing workflow inputs.

| Ordinary workflow payload | Default key ID |
|---|---|
| Caller workflow input and final result | `caller-workflow` |
| Compressed/encrypted handler's backing workflow inputs | `compressed-encrypted-1` |
| Encrypted handler's backing workflow inputs | `encrypted-1` |

These workflow payloads use AES-GCM without compression or endpoint metadata.
An async Nexus result uses the key and compression chain selected by its endpoint,
even when that differs from the backing workflow's input encryption.

For synchronous operations, the SDK supplies Nexus context when decoding inputs and
encoding results:

```text
Caller -- request encrypted with endpoint key --> Nexus endpoint --> Handler
Caller <-- response encrypted with endpoint key -- Nexus endpoint <-- Handler
```

When an asynchronous handler starts `EchoWorkflow`, the SDK includes the Nexus endpoint,
service, and operation in the workflow-start request. The server stores these fields in
the `WorkflowExecutionStarted` history event. The worker SDK restores the Nexus context
from that event when encoding the workflow's result, so `nexusCodec` can select the
endpoint's codec and key from the worker's local configuration. The workflow's input
still uses ordinary workflow context; the default codec encrypts it with that handler's
key 1.

```text
Caller -- encrypted request --> Endpoint --> Handler
                                               | decrypt with endpoint key
                                               | encrypt workflow input with default key
                                               v
                                         EchoWorkflow
                                               | worker decrypts input with default key
                                               | runs echo logic
                                               | encrypts result with endpoint key
                                               v
Caller <-- encrypted result -- Nexus completion
```

The caller starts all eight operations before waiting for their results. Each operation
keeps its own endpoint context, even when operations for different keys are in flight.
No custom interceptor or workflow-header propagation is required.

A workflow execution produces one serialized result, shared by all Nexus operations
attached to it. Operations that need differently encrypted results therefore need
separate workflow executions. This sample starts a backing workflow for each async
operation, using its Nexus request ID as the workflow ID suffix. Retrying the same
request uses the same workflow ID.

The outer encrypted payload includes `encoding`, `encryption-key-id`, and, for Nexus
payloads, `nexus-endpoint-name` metadata. For example:

```text
encoding: binary/nexus-aes-gcm
encryption-key-id: compressed-encrypted-1
nexus-endpoint-name: nexus-serialization-compressed-encrypted-1
```

These labels identify the key and endpoint; they do not contain the key material.
Ordinary workflow payloads have a key ID but no endpoint metadata or compression.

The AES-GCM codec and hard-coded keys are for this example. Use an established encryption
implementation and secure key storage for production.

## Prerequisites

Use Go SDK **v1.51.0 or later**, as configured in this repository, and a server that
supports Nexus serialization-context propagation to workflow-backed async results.

For local development, download the
[CLI `v1.9.3-nexus-serialization-context` archive](https://github.com/temporalio/cli/releases/tag/v1.9.3-nexus-serialization-context)
for your operating system and CPU architecture. Extract it and place the `temporal`
executable in the `samples-go` repository root. The commands below use `./temporal`
to run this build separately from your installed CLI.

## Run locally

Run the commands from the `samples-go` repository root. Check the CLI version first;
it should report `1.9.3-nexus-serialization-context`:

```shell
./temporal --version
```

Each worker polls a task queue with the same name as its namespace. Both endpoints
for a handler target that handler's namespace and task queue:

| Worker | Namespace and task queue |
|---|---|
| Caller | `nexus-serialization-caller` |
| Compressed/encrypted handler | `nexus-serialization-compressed-encrypted-handler` |
| Encrypted handler | `nexus-serialization-encrypted-handler` |

Start the server in its own terminal and leave it running:

```shell
./temporal server start-dev \
  --namespace nexus-serialization-caller \
  --namespace nexus-serialization-compressed-encrypted-handler \
  --namespace nexus-serialization-encrypted-handler
```

In another terminal, create two endpoints for each handler worker:

```shell
./temporal operator nexus endpoint create \
  --name nexus-serialization-compressed-encrypted-1 \
  --target-namespace nexus-serialization-compressed-encrypted-handler \
  --target-task-queue nexus-serialization-compressed-encrypted-handler
./temporal operator nexus endpoint create \
  --name nexus-serialization-compressed-encrypted-2 \
  --target-namespace nexus-serialization-compressed-encrypted-handler \
  --target-task-queue nexus-serialization-compressed-encrypted-handler
./temporal operator nexus endpoint create \
  --name nexus-serialization-encrypted-1 \
  --target-namespace nexus-serialization-encrypted-handler \
  --target-task-queue nexus-serialization-encrypted-handler
./temporal operator nexus endpoint create \
  --name nexus-serialization-encrypted-2 \
  --target-namespace nexus-serialization-encrypted-handler \
  --target-task-queue nexus-serialization-encrypted-handler
```

Run each worker in its own terminal and leave all three running alongside the server:

```shell
go run ./nexus-serialization-context/handler/compressed-encrypted-worker \
  -namespace nexus-serialization-compressed-encrypted-handler
```

```shell
go run ./nexus-serialization-context/handler/encrypted-worker \
  -namespace nexus-serialization-encrypted-handler
```

```shell
go run ./nexus-serialization-context/caller/worker \
  -namespace nexus-serialization-caller
```

Start the caller workflow in another terminal:

```shell
go run ./nexus-serialization-context/caller/starter \
  -namespace nexus-serialization-caller
```

### Output

The starter prints the workflow and run IDs, followed by a pair of results for each
endpoint. Omitting timestamps and replacing the generated IDs, the output is:

```text
Started workflow: ID=nexus-encryption-<workflow UUID> RunID=<run UUID>
nexus-serialization-compressed-encrypted-1: sync="Hello from Nexus" async="Hello from Nexus"
nexus-serialization-compressed-encrypted-2: sync="Hello from Nexus" async="Hello from Nexus"
nexus-serialization-encrypted-1: sync="Hello from Nexus" async="Hello from Nexus"
nexus-serialization-encrypted-2: sync="Hello from Nexus" async="Hello from Nexus"
```

All eight results should be `Hello from Nexus`. The starter exits when they are
available; the workers and server keep running. If only `Started workflow` appears,
the caller is still waiting for results.

The run creates five workflow executions: one caller and four async backing workflows.
The four sync operations return directly from their handlers.

## Inspect payloads

To inspect the encryption metadata for a backing workflow's input and result, follow
the async call through `nexus-serialization-compressed-encrypted-2`. These optional commands
require [jq](https://jqlang.org/).

Set `CALLER_WORKFLOW_ID` to the workflow ID printed by the starter. Find the async
request in its history; the handler prefixes that request ID with `nexus-echo-` to
form the backing workflow ID:

```shell
CALLER_WORKFLOW_ID='nexus-encryption-REPLACE-WITH-STARTER-ID'
BACKING_WORKFLOW_ID=$(
  ./temporal workflow show \
    --namespace nexus-serialization-caller \
    --workflow-id "$CALLER_WORKFLOW_ID" --output json |
  jq -r '.events[].nexusOperationScheduledEventAttributes |
    select(.endpoint == "nexus-serialization-compressed-encrypted-2" and .operation == "echo-async") |
    "nexus-echo-" + .requestId'
)
```

Inspect its start context and input/result metadata. JSON history represents metadata
values as base64; this command decodes those labels:

```shell
./temporal workflow show \
  --namespace nexus-serialization-compressed-encrypted-handler \
  --workflow-id "$BACKING_WORKFLOW_ID" --output json |
jq '.events[] |
  if .workflowExecutionStartedEventAttributes then
    .workflowExecutionStartedEventAttributes |
    {context: .propagatedNexusSerializationContext,
     inputMetadata: (.input.payloads[0].metadata | with_entries(.value |= @base64d))}
  elif .workflowExecutionCompletedEventAttributes then
    {resultMetadata: (.workflowExecutionCompletedEventAttributes.result.payloads[0].metadata |
      with_entries(.value |= @base64d))}
  else empty end'
```

The output should show:

- `context` contains endpoint `nexus-serialization-compressed-encrypted-2`, service
  `per-endpoint-encryption-service`, and operation `echo-async`.
- `inputMetadata` has key ID `compressed-encrypted-1` and no endpoint label: the
  backing workflow input uses the handler's ordinary workflow default.
- `resultMetadata` has key ID `compressed-encrypted-2` and endpoint label
  `nexus-serialization-compressed-encrypted-2`: the workflow result uses the propagated
  Nexus context.

Both payloads have encoding `binary/nexus-aes-gcm`. The compressed result's zlib metadata
is inside the encrypted payload.

## Test

Run the codec and caller workflow tests:

```shell
go test -race ./nexus-serialization-context/...
```
