# nexus-hello-api

## Getting started locally

### Get `temporal` CLI v0.14.0-nexus.0 to enable local development

```
curl -sSf https://temporal.download/cli.sh | sh -s -- --version v0.14.0-nexus.0 --dir .

./bin/temporal --version
```

### Spin up environment

#### Start temporal server

> HTTP port is required for Nexus communications

```
./bin/temporal server start-dev --http-port 7243 --dynamic-config-value system.enableNexus=true
```

### Initialize environment

In a separate terminal window

#### Create caller and target namespaces

```
./bin/temporal operator namespace create --namespace my-target-namespace
./bin/temporal operator namespace create --namespace my-caller-namespace
```

#### Create Nexus endpoint

```
./bin/temporal operator nexus endpoint create \
  --name my_nexus_endpoint_name \
  --target-namespace my-target-namespace \
  --target-task-queue my-handler-task-queue \
  --description-file ./service/description.md
```

## Getting started with a self-hosted service or Temporal Cloud

Nexus is currently available as
[pre-release](https://docs.temporal.io/evaluate/development-production-features/release-stages).

Self hosted users can [try Nexus
out](https://github.com/temporalio/temporal/blob/main/docs/architecture/nexus.md#trying-nexus-out) in single cluster
deployments with server version 1.25.0-rc.0 - **not meant for production use**.

Temporal Cloud users may reach out and open a support ticket to request access to the pre-release.

### Make Nexus calls across namespace boundaries

> Instructions apply for local development, for Temporal Cloud or a self hosted setups, supply the relevant [CLI
> flags](./options/cli.go) to properly set up the connection.

In separate terminal windows:

### Nexus handler worker

```
cd handler
go run ./worker \
    -target-host localhost:7233 \
    -namespace my-target-namespace
```

### Nexus caller worker

```
cd caller
go run ./worker \
    -target-host localhost:7233 \
    -namespace my-caller-namespace
```

### Start caller workflow

```
cd caller
go run ./starter \
    -target-host localhost:7233 \
    -namespace my-caller-namespace
```

### Output

which should result in:
```
2024/07/23 19:57:40 Workflow result: Nexus Echo 👋
2024/07/23 19:57:40 Started workflow WorkflowID nexus_hello_caller_workflow_20240723195740 RunID c9789128-2fcd-4083-829d-95e43279f6d7
2024/07/23 19:57:40 Workflow result: ¡Hola! Nexus 👋
```
