# Cloud Run Worker (identity + deployment versioning)

This sample demonstrates how to run a long-running Temporal Worker on
[Google Cloud Run](https://cloud.google.com/run) using the
[`cloudrun`](https://pkg.go.dev/go.temporal.io/sdk/contrib/gcp/cloudrun) contrib package to derive
the worker's **identity** and **Worker Deployment Version** from the Cloud Run environment.

At startup the worker reads its Cloud Run instance metadata once and uses it to:

- set a client **identity** of `<instanceID>@<revision>` (falling back to `<instanceID>@<name>`, then
  a bare `<instanceID>`), so each running instance is identifiable in the Temporal UI, and
- opt into [Worker Deployment Versioning](https://docs.temporal.io/worker-deployments) with a
  version of `(deploymentName, buildID) = (<name>, <revision>)` and a **PINNED** default versioning
  behavior. Cloud Run creates a new revision on every deploy, which makes it a natural build ID.

Cloud Run **worker pools** are the recommended deployment for Temporal workers: they are continuous,
pull-based background workloads with no request ingress, which is exactly the workload a Temporal
worker is. (The helper also works on Cloud Run **services** via `K_SERVICE` / `K_REVISION`.)

The sample registers a simple greeting Workflow and Activity, but the pattern applies to any
Workflow/Activity definitions.

## Files

| File | Description |
|------|-------------|
| `worker/main.go` | Long-running worker entry point -- reads Cloud Run metadata, applies the derived identity and PINNED deployment version, registers Workflows/Activities, and shuts down gracefully on SIGTERM |
| `starter/main.go` | Helper program to start a Workflow execution against the worker |
| `greeting/workflow.go` | Sample Workflow that executes a greeting Activity |
| `greeting/activity.go` | Sample Activity that returns a greeting string |
| `Dockerfile` | Builds the worker container image |

## Prerequisites

- A [Temporal Cloud](https://temporal.io/cloud) namespace (or a self-hosted Temporal cluster
  reachable from Cloud Run)
- The `gcloud` CLI configured for a project with the Cloud Run API enabled
- Go 1.26+

## Configure the Temporal connection

The worker and starter read the standard `TEMPORAL_ADDRESS`, `TEMPORAL_NAMESPACE`, and
`TEMPORAL_TASK_QUEUE` environment variables (defaulting to `localhost:7233`, `default`, and
`cloud-run-task-queue`). This sample uses a plaintext connection to keep it simple; add TLS or an API
key for Temporal Cloud as needed.

## Deploy to a Cloud Run worker pool

Deploy the worker straight from source. Cloud Run injects `CLOUD_RUN_WORKER_POOL` and
`CLOUD_RUN_REVISION`, which the helper reads to build the identity and deployment version:

```bash
gcloud run worker-pools deploy temporal-cloud-run-worker \
  --source . \
  --region=<REGION> \
  --set-env-vars TEMPORAL_ADDRESS=<namespace>.<account>.tmprl.cloud:7233,TEMPORAL_NAMESPACE=<namespace>.<account>,TEMPORAL_TASK_QUEUE=cloud-run-task-queue
```

`gcloud run worker-pools deploy` builds the container (using the provided `Dockerfile`), pushes it,
and rolls out a new **revision** to the worker pool. Because the worker pins workflows to its
deployment version, existing workflows keep running on the revision that started them until you move
them; new workflows start on the newest revision. Depending on your `gcloud` version, worker pools
may be under the `beta` track (`gcloud beta run worker-pools deploy ...`).

> Cloud Run does **not** expose the instance ID as an environment variable, so the worker fetches it
> from the GCP metadata server at `http://metadata.google.internal`. That server is reachable on
> Cloud Run worker pools and services but not when running locally, so the worker is meant to run on
> Cloud Run; use the starter below to drive it from your machine.

## Start a Workflow

With the same `TEMPORAL_*` values set locally, start a Workflow that the deployed worker will
execute:

```bash
TEMPORAL_ADDRESS=<...> TEMPORAL_NAMESPACE=<...> TEMPORAL_TASK_QUEUE=cloud-run-task-queue \
  go run ./starter
```

## Temporary module replace

`go.temporal.io/sdk/contrib/gcp/cloudrun` is not yet published, so `samples-go/go.mod` contains a
local `replace` pointing at an in-repo checkout of the module:

```
replace go.temporal.io/sdk/contrib/gcp/cloudrun => ../sdk-go-2/contrib/gcp/cloudrun
```

The helper compiles against the released `go.temporal.io/sdk`, so no replace is needed for the SDK
itself; adding this dependency does raise the module's `go` directive to 1.26. The replace makes
`go build`/`go run`/`go vet` work locally against the in-repo module (adjust the path if your SDK
checkout lives elsewhere). A container build's context is the samples repo, so it cannot reach the
sibling `../sdk-go-2` directory: the `Dockerfile` and `--source` deploy build only once the
module is published and this replace is removed. **This PR is a draft until then.**
