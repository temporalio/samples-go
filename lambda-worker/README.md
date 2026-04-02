# Lambda Worker

This sample demonstrates how to run a Temporal Worker inside an AWS Lambda function using
the [`lambdaworker`](https://pkg.go.dev/go.temporal.io/sdk/contrib/aws/lambdaworker)
contrib package. It includes optional OpenTelemetry instrumentation that exports traces
and metrics through AWS Distro for OpenTelemetry (ADOT).

The sample registers a simple greeting Workflow and Activity, but the pattern applies to
any Workflow/Activity definitions.

## Prerequisites

- A [Temporal Cloud](https://temporal.io/cloud) namespace (or a self-hosted Temporal
  cluster accessible from your Lambda)
- AWS CLI configured with permissions to create Lambda functions, IAM roles, and
  CloudFormation stacks
- mTLS client certificate and key for your Temporal namespace (place as `client.pem` and
  `client.key` in this directory). See more [here](https://docs.temporal.io/cloud/certificates).
- Go 1.25+

## Files

| File | Description |
|------|-------------|
| `main.go` | Lambda worker entry point -- configures the worker, registers Workflows/Activities, and starts the Lambda handler |
| `greeting/workflow.go` | Sample Workflow that executes a greeting Activity |
| `greeting/activity.go` | Sample Activity that returns a greeting string |
| `starter/main.go` | Helper program to start a Workflow execution against the Lambda worker |
| `temporal.toml` | Temporal client connection configuration (update with your namespace) |
| `otel-collector-config.yaml` | OpenTelemetry Collector sidecar configuration for ADOT |
| `deploy-lambda.sh` | Builds and deploys the Lambda function |
| `mk-iam-role.sh` | Creates the IAM role that allows Temporal Cloud to invoke the Lambda |
| `iam-role-for-temporal-lambda-invoke.yaml` | CloudFormation template for the IAM role |
| `extra-setup-steps` | Additional IAM and Lambda configuration for OpenTelemetry support |

## Setup

### 1. Configure Temporal connection

Edit `temporal.toml` with your Temporal Cloud namespace address and credentials. In production,
we'd recommend reading your credentials from a secret store, but to keep this example simple
the toml file defaults to reading them from keys bundled along with the Lambda code. You may
also set environment variables if you prefer not to use a file. For more information, see
our [docs on environment config](https://docs.temporal.io/develop/environment-configuration).

### 2. Create the IAM role

This creates the IAM role that Temporal Cloud assumes to invoke your Lambda function:

```bash
./mk-iam-role.sh <stack-name> <external-id> <lambda-arn>
```

The External ID is provided by Temporal Cloud in your namespace's serverless worker
configuration.

### 3. (Optional) Enable OpenTelemetry

If you want traces, metrics, and logs, run the extra setup to grant the Lambda role the
necessary permissions:

```bash
./extra-setup-steps <role-name> <function-name> <region> <account-id>
```

Update `otel-collector-config.yaml` with your function name and region as needed.

### 4. Deploy the Lambda function

```bash
./deploy-lambda.sh <function-name>
```

This cross-compiles for Linux/amd64, bundles the binary with configuration files, and
uploads to AWS Lambda.

### 5. Start a Workflow

Use the starter program to execute a Workflow on the Lambda worker, using
the same config file the Lambda uses for connecting to the server:

```bash
TEMPORAL_CONFIG_FILE=temporal.toml go run starter/main.go
```
