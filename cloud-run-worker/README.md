# Cloud Run Worker

This sample demonstrates how to run a long-running Temporal Worker on
[Google Cloud Run](https://cloud.google.com/run) using the
[`cloudrun`](https://pkg.go.dev/go.temporal.io/sdk/contrib/gcp/cloudrun) contrib
package. It includes OpenTelemetry instrumentation that exports traces and
metrics through a Google-Built OpenTelemetry Collector sidecar to Google Cloud
(Cloud Trace and Google Managed Service for Prometheus).

Cloud Run **worker pools** are the recommended deployment because Temporal
workers are continuous, pull-based background workloads with no request ingress.

The sample registers a simple greeting Workflow and Activity, but the pattern
applies to any Workflow/Activity definitions.

## Files

| File | Description |
|------|-------------|
| `worker/main.go` | Long-running worker entry point -- creates the plugin, dials Temporal, registers Workflows/Activities, and shuts down gracefully on SIGTERM |
| `starter/main.go` | Helper program to start a Workflow execution against the worker |
| `greeting/workflow.go` | Sample Workflow that executes a greeting Activity |
| `greeting/activity.go` | Sample Activity that returns a greeting string |
| `otel-collector-config.yaml` | Google-Built OpenTelemetry Collector sidecar configuration (metrics without batch, traces with batch, GCP resource detection) |
| `worker-pool.yaml` | Cloud Run worker pool manifest with the worker and collector sidecar, a startup probe, and container dependencies |
| `Dockerfile` | Builds the worker container image |
| `cloudbuild.yaml` | Cloud Build config that builds and pushes the image |

## Prerequisites

- A [Temporal Cloud](https://temporal.io/cloud) namespace (or a self-hosted
  Temporal cluster reachable from Cloud Run)
- The `gcloud` CLI configured for a project with Cloud Run, Cloud Build, Cloud
  Trace, and Monitoring APIs enabled
- Go 1.25+

## Configure the Temporal connection

The worker and starter load their connection from the environment via
[`envconfig`](https://pkg.go.dev/go.temporal.io/sdk/contrib/envconfig). Set the
standard `TEMPORAL_ADDRESS`, `TEMPORAL_NAMESPACE`, and `TEMPORAL_API_KEY`
environment variables (or a `temporal.toml`) before running.

## Run locally

Start the worker (it will export OTLP telemetry to `localhost:4317`; run a local
collector there if you want to see it, otherwise the export is a no-op):

```bash
cd cloud-run-worker
go run ./worker
```

In another terminal, start a workflow:

```bash
cd cloud-run-worker
go run ./starter
```

> **Note on the temporary module replace:** `go.temporal.io/sdk/contrib/gcp/cloudrun`
> is not yet published, so `samples-go/go.mod` contains a local
> `replace go.temporal.io/sdk/contrib/gcp/cloudrun => ../sdk-go/contrib/gcp/cloudrun`.
> This makes `go run`/`go build` work against the in-repo module, but a container
> build (whose build context is the samples repo) cannot reach `../sdk-go`. Remove
> the replace once the module is published; until then, build the image from a
> checkout that vendors or otherwise provides the module.

## Deploy to a Cloud Run worker pool

1. Build and push the worker image with Cloud Build:

   ```bash
   gcloud builds submit --config=cloud-run-worker/cloudbuild.yaml \
     --substitutions=_IMAGE=<REGION>-docker.pkg.dev/<PROJECT>/<REPO>/cloud-run-worker:latest .
   ```

2. Store the collector config and Temporal API key in Secret Manager:

   ```bash
   gcloud secrets create otel-collector-config --data-file=cloud-run-worker/otel-collector-config.yaml
   printf '%s' "<your-temporal-api-key>" | gcloud secrets create temporal-api-key --data-file=-
   ```

3. Edit `worker-pool.yaml` (image, region, Temporal connection) and deploy:

   ```bash
   gcloud beta run worker-pools replace cloud-run-worker/worker-pool.yaml --region=<REGION>
   ```

The worker pool runs the worker alongside the Google-Built OpenTelemetry
Collector sidecar. The worker's `dependsOn` on the collector and the collector's
health-check startup probe ensure the collector is ready before the worker
starts emitting telemetry.

## Shutdown lifecycle

Cloud Run sends `SIGTERM` and allows roughly ten seconds before `SIGKILL`. On
that signal the worker stops polling, closes the Temporal client, and calls
`plugin.Shutdown` with an eight-second deadline to flush buffered metrics and
traces before the process exits.

## Telemetry pipeline

Metrics are sent to `googlemanagedprometheus` **without** a collector `batch`
processor: batching can merge periodic and forced-shutdown snapshots of the same
cumulative series into one Google Monitoring write, which is rejected as
duplicate data. Traces are sent to `googlecloud` and may be batched
independently. The plugin exports metrics every 60 seconds by default.
