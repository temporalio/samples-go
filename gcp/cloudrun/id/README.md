# Cloud Run Id

Runs a Temporal Worker on a [Google Cloud Run](https://cloud.google.com/run) worker
pool using the [`cloudrun/id`](https://pkg.go.dev/go.temporal.io/sdk/contrib/gcp/cloudrun/id) contrib
plugin. Registering `CloudRunIDPlugin` on the client sets the Temporal client identity to
`<instanceID>@<revision>` (from Cloud Run instance metadata), so each instance is identifiable in the
Temporal UI. Cloud Run worker pools are the recommended way to run Temporal Workers: continuous,
pull-based background workloads with no request ingress.

The plugin fetches the instance ID from the GCP metadata server, which is unreachable locally, so the
worker is meant to run on Cloud Run; drive it from your machine with the starter.

## Prerequisites

- A [Temporal Cloud](https://temporal.io/cloud) namespace or reachable self-hosted cluster
- The `gcloud` CLI configured for a project with the Cloud Run API enabled
- Go 1.26+

The worker and starter read `TEMPORAL_ADDRESS`, `TEMPORAL_NAMESPACE`, and `TEMPORAL_TASK_QUEUE`
(defaulting to `localhost:7233`, `default`, and `cloud-run-task-queue`), using a plaintext
connection; add TLS or an API key for Temporal Cloud.

## Deploy and run

Deploy the worker (Cloud Run injects `CLOUD_RUN_WORKER_POOL` and `CLOUD_RUN_REVISION`):

```bash
gcloud run worker-pools deploy temporal-cloud-run-worker \
  --source . \
  --region=<REGION> \
  --set-env-vars TEMPORAL_ADDRESS=<namespace>.<account>.tmprl.cloud:7233,TEMPORAL_NAMESPACE=<namespace>.<account>,TEMPORAL_TASK_QUEUE=cloud-run-task-queue
```

Then start a workflow the deployed worker will execute:

```bash
go run ./gcp/cloudrun/id/starter
```

> `cloudrun/id` is unreleased, so `go.mod` uses a local `replace` to a sibling `sdk-go` checkout. A
> `--source`/container build cannot reach that path, so this stays a draft until the module ships.
