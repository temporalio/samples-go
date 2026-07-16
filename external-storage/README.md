# External Storage Sample

This sample demonstrates how to offload large workflow payloads to S3-compatible
object storage using the Temporal Go SDK's built-in `ExternalStorage` system,
combined with the SDK's zlib `PayloadCodec` so the payloads stored both inline
in Temporal and in S3 are compressed.

**Scenario:** A fulfillment center processes batches of shipping orders. The
workflow receives a small request (a batch ID and order count), then internally
calls a `FetchOrders` activity that returns the full list of orders with
customer records, line items, and handling notes. That list — several hundred
kilobytes even after compression — is passed to a second `ProcessOrders`
activity. Finally the workflow returns a small `BatchSummary` with totals.

Each payload is first compressed by the SDK's `NewZlibCodec`. The SDK then
checks the compressed size against the default 256 KiB threshold; payloads
still above it are stored in S3 and replaced inline with compact claim-check
references. The workflow's own input (`OrderBatchRequest`) and result
(`BatchSummary`) compress to a few hundred bytes and remain inline.

A mock S3 service ([s3-mock](./s3-mock)) is included so you can run the sample
locally without an AWS account or Docker. A codec server
([codec-server](./codec-server)) is included to retrieve and decompress payloads
on demand for the Temporal Web UI.

## Steps to run this sample

1. Run a [Temporal service](https://github.com/temporalio/samples-go/tree/main/#how-to-use).
   For local development, `temporal server start-dev` is the easiest option:
    ```
    > temporal server start-dev
    Temporal CLI 1.7.0 (Server 1.31.0, UI 2.49.1)

    Temporal Server:  localhost:7233
    Temporal UI:      http://localhost:8233
    Temporal Metrics: http://localhost:59980/metrics
    ```
2. In a separate terminal, run the mock S3 server. It listens on `:5000` and
   creates the `temporal-payloads` bucket. Leave it running.
    ```
    go run ./external-storage/s3-mock
    ```
3. In a separate terminal, run the worker:
    ```
    go run ./external-storage/worker
    ```
4. In a separate terminal, run the starter:
    ```
    go run ./external-storage/starter
    ```
   Example output:
    ```
    Starting workflow external-storage-20260515-120000 (batch_id=BATCH-20260515-120000, order_count=200)

    Batch BATCH-20260515-120000: 200 orders processed
      Total shipping cost: $28512.40
      Total weight:        19684.2 kg
      Avg delivery:        4.4 days
    ```
5. (Optional) Run the codec server in a fourth terminal:
    ```
    go run ./external-storage/codec-server
    ```
   The codec server exists purely so the Temporal Web UI (and CLI) can
   visualize this sample's payloads. It decompresses inline zlib payloads
   and resolves external storage references against the mock S3. The worker
   and starter don't talk to it; they apply the same codec + S3 driver
   themselves via the shared [data_converter.go](./data_converter.go), so the
   sample runs end-to-end without the codec server.

   In the Temporal Web UI (http://localhost:8233), open Settings → Data Encoder
   and set the Remote Codec Endpoint to `http://localhost:8081`. Reload the
   workflow page; the inline compressed payloads will be shown as readable
   JSON, and externally-stored payloads can be downloaded to fetch their
   actual content from the mock S3.

   The Web UI sends the namespace as the `X-Namespace` header on each request,
   so multi-namespace setups can dispatch by reading that header.

   | Endpoint | Behavior |
   | --- | --- |
   | `POST /encode` | Compress the payload, then offload to S3 if it exceeds the threshold. |
   | `POST /decode` | Retrieve any external storage references from S3, then decompress. Pass `?preserveStorageRefs=true` to leave references as-is. |
   | `POST /download` | All inputs must be storage references. Retrieves them from S3 and decompresses. |
6. Run `temporal workflow show` to see how payloads are stored. Use the
   workflow ID printed by the starter in step 4 (the `external-storage-<timestamp>`
   value on the `Starting workflow ...` line):
    ```
    temporal workflow show --workflow-id external-storage-<timestamp>
    ```
   A single run of this workflow produces six payloads. After zlib compression,
   four stay inline in Temporal history (a few hundred bytes to a few KiB
   each) and two are offloaded to S3 because they exceed the 256 KiB threshold
   even when compressed:

   | Payload | Type | Storage |
   | --- | --- | --- |
   | Workflow input | `OrderBatchRequest` | Inline |
   | `FetchOrders` input | `OrderBatchRequest` | Inline |
   | `FetchOrders` output | `[]Order` (~400 KiB) | **External (S3)** |
   | `ProcessOrders` input | `[]Order` (~400 KiB) | **External (S3)** |
   | `ProcessOrders` output | `[]ProcessedOrder` (~5 KiB) | Inline |
   | Workflow result | `BatchSummary` | Inline |

   The two offloaded payloads carry the same order list passed from one
   activity to the next, which is exactly what external storage is designed to
   handle: large blobs that flow between activities without bloating workflow
   history.

## How it works

The client's `DataConverter` is wrapped with the SDK's zlib codec, and
`client.Options.ExternalStorage` is set with the SDK's S3 driver
(`go.temporal.io/sdk/contrib/aws/s3driver`):

```go
driver, _ := s3driver.NewDriver(s3driver.Options{
    Client: awssdkv2.NewClient(s3Client),
    Bucket: s3driver.StaticBucket(externalstorage.S3Bucket),
})
client.Dial(client.Options{
    DataConverter: converter.NewCodecDataConverter(
        converter.GetDefaultDataConverter(),
        converter.NewZlibCodec(converter.ZlibCodecOptions{AlwaysEncode: true}),
    ),
    ExternalStorage: converter.ExternalStorage{
        Drivers: []converter.StorageDriver{driver},
    },
})
```

While encoding a payload, the SDK:

1. Serializes the Go value to a `Payload`.
2. Runs the zlib codec to compress the payload bytes.
3. Checks the compressed size against `PayloadSizeThreshold` (default: 256 KiB).
4. If still above the threshold, stores the compressed bytes in S3 via
   the SDK's `s3driver` and replaces the inline payload with a claim-check
   reference.

While decoding a payload, the SDK reverses these steps, transparently retrieving
from S3 and decompressing as needed.

The worker, the starter, and the codec server must use the **same** codec and
external storage configuration so each side can read what the other wrote. In
this sample, the shared wiring lives in
[data_converter.go](./data_converter.go) for the worker and starter, and is
mirrored in [codec-server/main.go](./codec-server/main.go) for the codec
server.

## Codec server

The codec server is built directly on top of the SDK's
`converter.NewPayloadHTTPHandler`, which implements the `/encode`, `/decode`,
and `/download` endpoints with full external storage support. The sample adds
two thin layers around it:

- A **namespace dispatcher** that picks a per-namespace handler by inspecting
  the `X-Namespace` header sent by the Temporal Web UI and CLI. Only `"default"`
  is configured here, but the same map can host other namespaces with their own
  codec chains and storage backends.
- A **CORS middleware** that allows the Web UI origin to call the codec
  server.

### Which codec server should I build?

The SDK offers two HTTP handler constructors, intended for different consumers:

- **`converter.NewPayloadHTTPHandler` (used in this sample)** — for the
  Temporal Web UI and CLI. It serves `/encode`, `/decode`, and `/download`,
  and understands external storage references so users can inspect and
  retrieve offloaded payloads from a browser. It is **not** meant to be
  plugged into an SDK client's `DataConverter`.
- **`converter.NewPayloadCodecHTTPHandler`** — for SDK clients that want to
  offload their codec chain to a remote service (e.g. to centralize key
  material or codec configuration). It serves the older `/encode` + `/decode`
  protocol without external storage semantics. Clients consume it via
  `converter.NewRemoteDataConverter` or `converter.NewRemotePayloadCodec`,
  both of which expect that protocol.

If you need both — codec-as-a-service for clients/workers **and** Web UI / CLI
visualization that includes external storage — run one handler of each kind,
bound at distinct routes (e.g. `/ui/encode`, `/ui/decode`, `/ui/download` for
the Web UI handler and `/codec/encode`, `/codec/decode` for the client
handler) so the two protocols don't collide on `/encode` and `/decode`.
Configure each consumer with the URL prefix that matches its handler.
