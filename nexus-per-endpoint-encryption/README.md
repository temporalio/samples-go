# Nexus per-endpoint encryption

A caller invokes synchronous and workflow-backed asynchronous echo operations through
four Nexus endpoints. Each handler serves two endpoints and selects a distinct AES-GCM
key for each using `converter.NexusSerializationContext.Endpoint`.

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

The caller starts all eight calls before waiting for their results, then returns four
pairs of sync/async replies. The starter prints all eight replies.

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
labels, not secret keys. Decode checks them before AES-GCM authentication. Real keys
must come from a secure key store; the hard-coded keys are only for this local sample.

## Prerequisites

Use Go SDK **v1.51.0 or later**, as configured in this repository, and a server supporting
Nexus serialization-context propagation to workflow-backed async results. The
[SDK v1.51.0 release](https://github.com/temporalio/sdk-go/releases/tag/v1.51.0)
includes that SDK support.

For local development, download
[CLI `v1.9.3-nexus-serialization-context`](https://github.com/temporalio/cli/releases/tag/v1.9.3-nexus-serialization-context).
Its bundled server supports this feature. Keep the binary separate from your stable
CLI; commands below use `./temporal` from the repository root. This specifies the
feature-enabled build used to verify the sample, rather than a minimum stable server
version.

## Run locally

The caller and each handler run in separate namespaces. Each task queue has the same
name as its namespace:

| Worker | Namespace and task queue |
|---|---|
| Caller | `nexus-serialization-caller` |
| Compressed/encrypted handler | `nexus-serialization-compressed-encrypted-handler` |
| Encrypted handler | `nexus-serialization-encrypted-handler` |

Start the server:

```shell
./temporal server start-dev \
  --namespace nexus-serialization-caller \
  --namespace nexus-serialization-compressed-encrypted-handler \
  --namespace nexus-serialization-encrypted-handler
```

Create the endpoints:

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
go run ./nexus-per-endpoint-encryption/handler/compressed-encrypted-worker \
  -namespace nexus-serialization-compressed-encrypted-handler
```

```shell
go run ./nexus-per-endpoint-encryption/handler/encrypted-worker \
  -namespace nexus-serialization-encrypted-handler
```

```shell
go run ./nexus-per-endpoint-encryption/caller/worker \
  -namespace nexus-serialization-caller
```

Start the caller:

```shell
go run ./nexus-per-endpoint-encryption/caller/starter \
  -namespace nexus-serialization-caller
```

Each endpoint prints `sync="Hello from Nexus" async="Hello from Nexus"`.
Use the existing Nexus sample flags such as `-target-host` to connect elsewhere.

## Test

```shell
go test -race ./nexus-per-endpoint-encryption/...
```

The tests check the sample's converter configuration and caller orchestration. Codec
tests supply serialization context explicitly; they do not prove SDK/server context
propagation. They do not start or download a server.
