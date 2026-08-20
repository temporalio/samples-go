# Nexus per-endpoint encryption

This sample extends the [getting started Nexus sample](../nexus) with a context-aware payload codec. The codec uses
`converter.NexusSerializationContext.Endpoint` to select a different AES-GCM key for each Nexus endpoint.
The serialization context also identifies the Nexus service and operation, so a codec can instead select keys by
service, by operation, or by any combination of endpoint, service, and operation.

Per-endpoint key selection has an important limitation for workflow-backed operations. Multiple Nexus operations can
attach completion callbacks to the same workflow execution, including operations invoked through different endpoints.
Because that workflow execution produces one serialized result shared by all attached operations, the result cannot be
encrypted separately with a different key for each endpoint. Applications using this pattern must ensure that every
operation attached to the same workflow execution resolves to the same key (or avoid attaching operations from
different key domains to that execution). Application that must should use a wrapper workflow to decouple the results.


The caller workflow starts four Nexus operations before waiting for their results:

- a synchronous operation through `nexus-encryption-endpoint-a`
- a workflow-backed asynchronous operation through `nexus-encryption-endpoint-a`
- a synchronous operation through `nexus-encryption-endpoint-b`
- a workflow-backed asynchronous operation through `nexus-encryption-endpoint-b`

Ordinary workflow payloads are left unchanged. Nexus operation inputs and outputs, including the asynchronous handler
workflow input and result, are encrypted with the key assigned to the endpoint.

This sample uses unreleased Go SDK support for `converter.NexusSerializationContext`. While developing it alongside a
checkout of `sdk-go`, create a local workspace from the `samples-go` directory:

```shell
go work init .
go work use ../sdk-go
```

The workspace file is only needed until `samples-go` depends on an SDK release containing the feature.

## Run the sample

Start a local Temporal server:

```shell
temporal server start-dev
```

Create the caller and handler namespaces:

```shell
temporal operator namespace create --namespace nexus-encryption-caller
temporal operator namespace create --namespace nexus-encryption-handler
```

Create two endpoints that route to the same handler worker:

```shell
temporal operator nexus endpoint create \
  --name nexus-encryption-endpoint-a \
  --target-namespace nexus-encryption-handler \
  --target-task-queue nexus-encryption-handler-task-queue

temporal operator nexus endpoint create \
  --name nexus-encryption-endpoint-b \
  --target-namespace nexus-encryption-handler \
  --target-task-queue nexus-encryption-handler-task-queue
```

Run the handler and caller workers in separate terminals:

```shell
go run ./nexus-per-endpoint-encryption/handler/worker \
  -target-host localhost:7233 \
  -namespace nexus-encryption-handler
```

```shell
go run ./nexus-per-endpoint-encryption/caller/worker \
  -target-host localhost:7233 \
  -namespace nexus-encryption-caller
```

Start the caller workflow:

```shell
go run ./nexus-per-endpoint-encryption/caller/starter \
  -target-host localhost:7233 \
  -namespace nexus-encryption-caller
```

The endpoint keys in this sample are hard-coded for clarity. Production applications should load keys from a secure
key-management system and use stable key identifiers to support rotation.
