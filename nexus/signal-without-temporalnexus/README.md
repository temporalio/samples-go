# Signal from Nexus without temporalnexus

This sample demonstrates that a Nexus operation handler can signal a workflow without importing
`go.temporal.io/sdk/temporalnexus`.

The operation is registered directly with `nexus.NewSyncOperation`, captures a normal
`client.Client`, and calls `SignalWorkflow(ctx, ...)` using the Nexus handler context.

## Run

Start a local Temporal server:

```bash
temporal server start-dev
```

Create the Nexus endpoint:

```bash
temporal operator nexus endpoint create \
  --name nexus-signal-without-temporalnexus-endpoint \
  --target-namespace default \
  --target-task-queue nexus-signal-without-temporalnexus-task-queue
```

Run the sample:

```bash
go run ./nexus/signal-without-temporalnexus
```

Expected output includes:

```text
Caller workflow result: signaled workflow "nexus-signal-receiver-..."
Receiver workflow result: signal sent from a raw Nexus operation
```
