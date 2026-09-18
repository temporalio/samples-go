# Cloud Run worker (OpenTelemetry)

Runs a Temporal Worker on a [Google Cloud Run](https://cloud.google.com/run) worker pool
using the [`cloudrun/otel`](https://pkg.go.dev/go.temporal.io/sdk/contrib/gcp/cloudrun/otel) contrib
plugin. The plugin exports traces and metrics over OTLP to a Google-Built OpenTelemetry Collector
sidecar (`worker-pool.yaml`), which forwards them to Cloud Trace and Google Managed Service for
Prometheus. Cloud Run worker pools are the recommended way to run Temporal workers: continuous,
pull-based background workloads with no request ingress.

## Prerequisites

- A [Temporal Cloud](https://temporal.io/cloud) namespace or reachable self-hosted cluster
- The `gcloud` CLI configured for a project with the Cloud Run, Cloud Build, Cloud Trace, and
  Monitoring APIs enabled
- Go 1.26+

The worker and starter load their connection from the environment via
[`envconfig`](https://pkg.go.dev/go.temporal.io/sdk/contrib/envconfig): set `TEMPORAL_ADDRESS`,
`TEMPORAL_NAMESPACE`, and `TEMPORAL_API_KEY` (or a `temporal.toml`).

## Run locally

The worker exports OTLP to `localhost:4317`; without a local collector there the export is a no-op.

```bash
go run ./gcp/cloudrun/otel/worker    # in one terminal
go run ./gcp/cloudrun/otel/starter   # in another
```

## Deploy to a Cloud Run worker pool

```bash
# 1. Build and push the image.
gcloud builds submit --config=gcp/cloudrun/otel/cloudbuild.yaml \
  --substitutions=_IMAGE=<REGION>-docker.pkg.dev/<PROJECT>/<REPO>/cloud-run-worker:latest .

# 2. Store the collector config and API key in Secret Manager.
gcloud secrets create otel-collector-config --data-file=gcp/cloudrun/otel/otel-collector-config.yaml
printf '%s' "<temporal-api-key>" | gcloud secrets create temporal-api-key --data-file=-

# 3. Edit worker-pool.yaml (image, region, Temporal connection) and deploy.
gcloud beta run worker-pools replace gcp/cloudrun/otel/worker-pool.yaml --region=<REGION>
```

On SIGTERM the worker stops polling, closes the client, and flushes telemetry via `plugin.Shutdown`.
Metrics go to `googlemanagedprometheus` without a collector `batch` processor, because batching can
merge cumulative-series snapshots into a duplicate Monitoring write; traces go to `googlecloud`.

> `cloudrun/otel` and its `opentelemetry/otlpworker` dependency are unreleased, so `go.mod` uses local
> `replace` directives to a sibling `sdk-go` checkout. A `--source`/container build cannot reach that
> path, so this stays a draft until the modules ship.
