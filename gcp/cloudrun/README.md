# Cloud Run worker

Runs a Temporal Worker on a [Google Cloud Run](https://cloud.google.com/run) worker pool using two
GCP Cloud Run contrib plugins together.

- [`cloudrun/id`](https://pkg.go.dev/go.temporal.io/sdk/contrib/gcp/cloudrun/id) sets the Temporal
  client identity to `<instanceID>@<revision>` from Cloud Run instance metadata, so each instance is
  identifiable in the Temporal UI.
- [`cloudrun/otel`](https://pkg.go.dev/go.temporal.io/sdk/contrib/gcp/cloudrun/otel) exports SDK
  traces and metrics over OTLP to a Google-Built OpenTelemetry Collector sidecar (`worker-pool.yaml`),
  which forwards them to Cloud Trace and Google Managed Service for Prometheus.

## Prerequisites

- A [Temporal Cloud](https://temporal.io/cloud) namespace or reachable self-hosted cluster
- The `gcloud` CLI configured for a project with the Cloud Run, Cloud Build, Cloud Trace, and
  Monitoring APIs enabled
- Go 1.26+

The worker and starter load their connection from the environment via
[`envconfig`](https://pkg.go.dev/go.temporal.io/sdk/contrib/envconfig): set `TEMPORAL_ADDRESS`,
`TEMPORAL_NAMESPACE`, and `TEMPORAL_API_KEY` (or a `temporal.toml`).

## Deploy to a Cloud Run worker pool

```bash
# 1. Create the Artifact Registry repo (once), then build and push the image.
gcloud artifacts repositories create <REPO> --repository-format=docker --location=<REGION>
gcloud builds submit --config=gcp/cloudrun/cloudbuild.yaml \
  --substitutions=_IMAGE=<REGION>-docker.pkg.dev/<PROJECT>/<REPO>/cloud-run-worker:latest .

# 2. Store the collector config and API key in Secret Manager.
gcloud secrets create otel-collector-config --data-file=gcp/cloudrun/otel-collector-config.yaml
printf '%s' "<temporal-api-key>" | gcloud secrets create temporal-api-key --data-file=-

# 3. Edit worker-pool.yaml (image, region, Temporal connection) and deploy.
gcloud beta run worker-pools replace gcp/cloudrun/worker-pool.yaml
```

Set `<SERVICE_ACCOUNT>` in `worker-pool.yaml` to a service account holding `roles/monitoring.metricWriter`, `roles/telemetry.tracesWriter`, and `roles/secretmanager.secretAccessor`.

`replace` prints `Done.` once the pool is ready and the worker starts polling.

## Start a workflow

The Cloud Run Id plugin needs the GCP metadata server, so the worker runs on Cloud Run. Once it is
deployed, drive it from your machine:

```bash
go run ./gcp/cloudrun/starter
```

This prints `Workflow result: Hello, Cloud Run Worker!`, confirming the deployed worker ran the task.

On SIGTERM the worker stops polling, closes the client, and flushes telemetry via `plugin.Shutdown`.
Metrics go to Managed Service for Prometheus without a collector `batch` processor, because batching
can merge cumulative-series snapshots into a duplicate Monitoring write; traces go to Cloud Trace.
