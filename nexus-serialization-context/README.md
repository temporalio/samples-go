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

These four endpoint keys are distinct 128-bit AES keys. The AES-GCM codec is implemented in this sample.
Compression uses the SDK's zlib codec. Nexus supplies the endpoint context used to select them.

`nexusCodec` implements `converter.PayloadCodecWithSerializationContext` using an
endpoint-to-codec map and a default codec for ordinary workflow payloads. To implement this
pattern, wrap the codec with `converter.NewCodecDataConverter` and configure it through
`client.Options.DataConverter` on the caller and handler clients. Both sides need
matching keys and encodings for each endpoint.

In this sample, the caller uses a separate key for its own input and final result, keeping those
payloads independent of the keys shared with Nexus handlers. Each handler reuses its
key 1 for backing workflow inputs.

| Ordinary workflow payload | Default key ID |
|---|---|
| Caller workflow input and final result | `caller-workflow` |
| Compressed/encrypted handler's backing workflow inputs | `compressed-encrypted-1` |
| Encrypted handler's backing workflow inputs | `encrypted-1` |

These workflow payloads use AES-GCM without compression or endpoint metadata.
An async Nexus result will use the key and compression chain selected by its endpoint,
even when that differs from the backing workflow's input encryption.

The SDK supplies Nexus context when encoding and decoding operation requests and
results. For workflow-backed asynchronous operations, it preserves that context in
workflow history and restores it when encoding the result. A call through endpoint 2
can therefore return a result encrypted with key 2, as the caller expects, even though
the backing workflow's input uses the handler's default key 1.

The caller starts all eight operations before waiting for their results. Each operation
keeps its own endpoint context, even when operations for different keys are in flight.
No custom interceptor or workflow-header propagation is required.

Note that a workflow execution produces one serialized result, shared by all Nexus operations
attached to it. Operations that need differently encrypted results therefore need
separate workflow executions. This sample starts a backing workflow for each async
operation, using its Nexus request ID as the workflow ID suffix. As a result, retrying 
the same request uses the same workflow ID.

The outer encrypted payload includes `encoding`, `encryption-key-id`, and, for Nexus
payloads, `nexus-endpoint-name` metadata. For example:

```text
encoding: binary/nexus-aes-gcm
encryption-key-id: compressed-encrypted-1
nexus-endpoint-name: nexus-serialization-compressed-encrypted-1
```

These labels identify the key and endpoint; they do not contain the key material.
Ordinary workflow payloads have a key ID but no endpoint metadata or compression.

The AES-GCM codec and hard-coded keys are for this example. For production, use an established encryption
implementation and secure key storage.

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

All eight results should be `Hello from Nexus`. If only `Started workflow` appears,
the caller is still waiting for results.

The run creates five workflow executions: one caller and four async backing workflows.
The four sync operations return directly from their handlers.

## Test

Run the codec and caller workflow tests:

```shell
go test -race ./nexus-serialization-context/...
```
