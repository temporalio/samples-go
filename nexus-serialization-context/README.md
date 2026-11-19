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

Each endpoint uses a different 128-bit AES key. Compression runs before encryption,
and decoding decrypts before decompressing. AES-GCM authenticates the ciphertext.
The AES-GCM codec is implemented in this sample. Compression uses the SDK's zlib codec.
Nexus supplies the endpoint context used to select a codec. The context also identifies
the service and operation, which can be used to select keys.

## Adapt this pattern

`nexusCodec` implements `converter.PayloadCodecWithSerializationContext`. It maps
endpoints to codecs and uses a default codec for ordinary workflow payloads. Wrap it
with `converter.NewCodecDataConverter` and set `client.Options.DataConverter` on the
caller and handler clients. See [encryption/converter.go](encryption/converter.go) for the key maps and
[handler/app.go](handler/app.go) for client setup.

Because AES-GCM uses symmetric encryption, the caller and handler share a secret key
for each endpoint. This sample uses that key for both requests and results.
Both sides also need the same payload format, including whether to compress.

The caller uses a separate key for its workflow input and final result.
For this example, each handler reuses its endpoint-1 key for backing workflow inputs.
If your application already has a workflow codec, keep it as the default so it
can still read existing histories.

| Ordinary workflow payload | Default key ID |
|---|---|
| Caller workflow input and final result | `caller-workflow` |
| Compressed/encrypted handler's backing workflow inputs | `compressed-encrypted-1` |
| Encrypted handler's backing workflow inputs | `encrypted-1` |

These workflow payloads use AES-GCM without compression or endpoint metadata.

On decode, the AES-GCM codec passes other encodings through for ordinary decoding.
For AES-GCM payloads, it checks the key ID against the codec selected by context and
verifies the ciphertext before returning the decoded value.

The SDK supplies Nexus context for operation requests and results. For workflow-backed
async operations, the server stores the context in workflow history and the SDK uses it
to select the result's key and compression. To see the two keys in one execution, call
`nexus-serialization-encrypted-2` and check the backing workflow's history. Its input
records `encrypted-1` and its result records `encrypted-2`.

The caller starts all eight operations before waiting for their results. Each operation
keeps its own endpoint context while calls using different keys run concurrently.
No custom interceptor or workflow-header propagation is required.

Each workflow execution has one serialized result, shared by all Nexus operations
attached to it. Operations that need results encrypted with different keys must use
separate workflow executions. This sample starts one backing workflow per async
operation, using the Nexus request ID as its workflow ID suffix. Retrying the same
request reuses that workflow ID.

The outer encrypted payload includes `encoding`, `encryption-key-id`, and, for Nexus
payloads, `nexus-endpoint-name` metadata. For example:

```text
encoding: binary/nexus-aes-gcm
encryption-key-id: compressed-encrypted-1
nexus-endpoint-name: nexus-serialization-compressed-encrypted-1
```

The metadata identifies the key and endpoint. The key material stays in the codec
configuration.

The AES-GCM codec and hard-coded keys are for this example. For production, use an
established encryption implementation and secure key storage.

## Prerequisites

Use Go SDK **v1.51.0 or later**, as configured in this repository, and a server that
supports Nexus serialization-context propagation to workflow-backed async results.

For local development, download the
[CLI `v1.9.3-nexus-serialization-context` archive](https://github.com/temporalio/cli/releases/tag/v1.9.3-nexus-serialization-context)
for your operating system and CPU architecture. Extract it and place the `temporal`
executable in the `samples-go` repository root. The commands below use `./temporal`
to run this build separately from your installed CLI.

## Run locally

Run the commands from the `samples-go` repository root. Check the CLI version first.
It should report `1.9.3-nexus-serialization-context`:

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

Start the dev server in its own terminal:

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

Run each worker in its own terminal:

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

Expected output (timestamps omitted):

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
