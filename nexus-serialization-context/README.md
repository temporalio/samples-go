# Nexus serialization context

Use Nexus serialization context to select encryption keys and compression by endpoint
for synchronous and workflow-backed asynchronous operations. A caller invokes echo
operations through four Nexus endpoints. Each handler serves two endpoints and selects
a distinct AES-GCM key for each using `converter.NexusSerializationContext.Endpoint`.

| Handler worker | Endpoint | Key ID | Nexus payload encoding |
|---|---|---|---|
| Compressed/encrypted | `nexus-serialization-compressed-encrypted-1` | `compressed-encrypted-1` | zlib, then AES-GCM |
| Compressed/encrypted | `nexus-serialization-compressed-encrypted-2` | `compressed-encrypted-2` | zlib, then AES-GCM |
| Encrypted | `nexus-serialization-encrypted-1` | `encrypted-1` | AES-GCM |
| Encrypted | `nexus-serialization-encrypted-2` | `encrypted-2` | AES-GCM |

Each worker uses the data converter configured on its client. The sample's Nexus codec
implements `converter.PayloadCodecWithSerializationContext` to choose an endpoint's
codec. Context-bound copies share read-only configuration, so concurrent calls through
different endpoints do not change each other's key selection. Compression uses the
SDK's zlib codec; AES-GCM is implemented in this sample using Go's standard library,
following the repository's [encryption sample](../encryption).

The reusable pattern is endpoint selection and context binding in
[`encryption/converter.go`](encryption/converter.go). The AES-GCM codec and hard-coded
keys are demonstration code for local use. For production, adapt that pattern to an
established encryption implementation and key-management service, such as
[Tink Go](https://github.com/tink-crypto/tink-go) and its
[AWS KMS integration](https://github.com/tink-crypto/tink-go-awskms), and keep key
material out of source code.

The caller starts all eight calls before waiting for their results, then returns four
pairs of sync/async replies. The starter prints all eight replies.

This sample demonstrates successful synchronous and workflow-backed asynchronous
operations. The [broader SDK feature](https://github.com/temporalio/sdk-go/releases/tag/v1.51.0)
also supports failures, workflow updates, and standalone activity results and
heartbeat details.

## Prerequisites

Use Go SDK **v1.51.0 or later**, as configured in this repository, and a server supporting
Nexus serialization-context propagation to workflow-backed async results. The
[SDK v1.51.0 release](https://github.com/temporalio/sdk-go/releases/tag/v1.51.0)
includes that SDK support.

For local development, download the
[CLI `v1.9.3-nexus-serialization-context` archive](https://github.com/temporalio/cli/releases/tag/v1.9.3-nexus-serialization-context)
for your operating system and CPU architecture. Extract the archive and place its
`temporal` executable in the `samples-go` repository root. Keep this binary separate
from your stable CLI installation; leave your `PATH` unchanged. Its bundled server
supports this feature. This is the build used to verify the sample, rather than a
minimum stable server version.

## Run locally

Run every command below from the `samples-go` repository root, including commands in
new terminals. Confirm that the extracted binary reports
`1.9.3-nexus-serialization-context`:

```shell
./temporal --version
```

The caller and each handler run in separate namespaces. Each task queue has the same
name as its namespace:

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

In another terminal, create the endpoints:

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

Start the caller in a fifth terminal (you can reuse the terminal that created the endpoints):

```shell
go run ./nexus-serialization-context/caller/starter \
  -namespace nexus-serialization-caller
```

The starter prints the caller's workflow and run IDs, followed by four result lines.
Ignoring the timestamp prefixes, the output looks like this (your IDs will differ):

```text
Started workflow: ID=nexus-encryption-<workflow UUID> RunID=<run UUID>
nexus-serialization-compressed-encrypted-1: sync="Hello from Nexus" async="Hello from Nexus"
nexus-serialization-compressed-encrypted-2: sync="Hello from Nexus" async="Hello from Nexus"
nexus-serialization-encrypted-1: sync="Hello from Nexus" async="Hello from Nexus"
nexus-serialization-encrypted-2: sync="Hello from Nexus" async="Hello from Nexus"
```

Success means all four endpoints returned both replies: eight completed operations
in total. One caller workflow coordinates the eight operations; each async operation
starts a backing workflow, for five workflow executions in total. Sync operations
return directly from their handlers. The starter then exits; the three workers and
server keep running. If only the `Started workflow` line appears, the caller is still
waiting for results.

## Verify the result

The echo replies confirm that the calls completed. To see which keys the async path
used, inspect the backing workflow for `nexus-serialization-compressed-encrypted-2`.
The optional commands below also require [jq](https://jqlang.org/).

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

Check that:

- `context` contains endpoint `nexus-serialization-compressed-encrypted-2`, service
  `per-endpoint-encryption-service`, and operation `echo-async`.
- `inputMetadata` has key ID `compressed-encrypted-1` and no endpoint label: the
  backing workflow input uses the handler's ordinary workflow default.
- `resultMetadata` has key ID `compressed-encrypted-2` and endpoint label
  `nexus-serialization-compressed-encrypted-2`: the workflow result uses the propagated
  Nexus context.

Both payloads have outer encoding `binary/nexus-aes-gcm`. Compression is inside that
encrypted envelope, so these labels alone do not prove compression. The compression
test below checks that compressed endpoint payloads round-trip and are smaller for
repeated input.

## Workflow encryption and async results

Ordinary workflow payloads have workflow context, rather than Nexus context. Each
worker's converter has a default encryption key for those payloads:

| Ordinary workflow payload | Default key ID |
|---|---|
| Caller workflow input and final result | `caller-workflow` |
| Compressed/encrypted handler's backing workflow inputs | `compressed-encrypted-1` |
| Encrypted handler's backing workflow inputs | `encrypted-1` |

These defaults use AES-GCM without compression or endpoint metadata. Both endpoints
for each handler use the same default for backing workflow inputs. The caller's fifth
key is independent of the four keys used for Nexus exchanges.

For a synchronous operation, the handler decodes the request and encodes the response
using the endpoint's codec:

```text
Caller -- request encrypted with endpoint key --> Nexus endpoint --> Handler
Caller <-- response encrypted with endpoint key -- Nexus endpoint <-- Handler
```

For an asynchronous operation, the handler starts `EchoWorkflow`. The SDK sends Nexus
endpoint/service/operation context in the start request; the server stores it in the
workflow-start history event. The worker restores that context to encode the result:

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

For endpoint 2, the backing workflow input therefore uses key 1, while its result uses
key 2 (and compression for the compressed/encrypted handler). No custom interceptor or
workflow header is needed. The server stores encrypted payloads; keys remain local.

Each async operation starts its own backing workflow, identified by its Nexus request
ID. Retrying that request uses the same workflow ID. Do not attach operations from
different endpoint key domains to the same existing workflow: one workflow produces
one serialized result, shared by its attached operations.

Encrypted payload metadata contains `encoding: binary/nexus-aes-gcm` and
`encryption-key-id`. Nexus payloads also contain `nexus-endpoint-name`. These are public
labels, not secret keys. Decode checks them before AES-GCM authentication.

## Test

```shell
go test -race ./nexus-serialization-context/...
```

The tests check the sample's converter configuration and caller orchestration. Codec
tests independently verify the expected raw keys, retained context bindings, fresh
ciphertext for repeated input, and rejection of altered or truncated ciphertext.
They supply serialization context explicitly; they do not prove SDK/server context
propagation. They do not start or download a server.
