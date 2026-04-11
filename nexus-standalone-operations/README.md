This sample demonstrates how to use Standalone Nexus Operations (executing Nexus operations directly from client code without wrapping them in a Workflow).
It shows both sync and async (workflow-backed) operations, and how to use the `ListNexusOperations` and `CountNexusOperations` APIs.

## Note: Standalone Nexus operations require a server version that supports this feature. Use the dev server build at https://github.com/temporalio/cli/releases/tag/v1.7.2-standalone-nexus-operations.

### Steps to run this sample (with expected output):
1) Run the [Temporal dev server build that supports standalone Nexus operations](https://github.com/temporalio/cli/releases/tag/v1.7.2-standalone-nexus-operations). (If you are going to run locally, you will want to start it in another terminal; this command is blocking and runs until it receives a SIGINT (Ctrl + C) command.)

Start the dev server with the dynamic config flags required for standalone Nexus operations:

```bash
temporal server start-dev \
  --dynamic-config-value "nexusoperation.enableStandalone=true" \
  --dynamic-config-value "history.enableChasmCallbacks=true"
```

You should see a line about the CLI, Server and UI versions, and one line each for the Server URL, UI URL and Metrics endpoint. It should look something like this:

```bash
Temporal CLI 1.7.2-standalone-nexus-operations (Server 1.32.0-155.0, UI 2.49.1)

Temporal Server:  localhost:7233
Temporal UI:      http://localhost:8233
Temporal Metrics: http://localhost:61951/metrics
```

2) Create a Nexus endpoint that routes to the worker's task queue. In a second terminal, run:
```bash
temporal operator nexus endpoint create \
  --name nexus-standalone-operations-endpoint \
  --target-namespace default \
  --target-task-queue nexus-standalone-operations
```

1) Then run the following command to start the worker. The worker is a blocking process that runs until it receives a SIGINT (Ctrl + C) command.
```bash
go run nexus-standalone-operations/worker/main.go
```

You should see the following log line:
1. Starting the Worker with Namespace `default`, and TaskQueue `nexus-standalone-operations` and it will list the WorkerID for the created worker.

For example:
```bash
2026/05/21 08:59:49 INFO  Started Worker Namespace default TaskQueue nexus-standalone-operations WorkerID 71172
```

> [!NOTE]
> Timestamps and IDs will differ on your machine.

4) In a third terminal, run the following command to start the example:
```bash
go run nexus-standalone-operations/starter/main.go
```

You should see something similar to the following output:

```bash
2026/05/21 09:00:30 Started Echo operation OperationID nexus-standalone-echo-op
2026/05/21 09:00:30 Echo result: hello
2026/05/21 09:00:30 Started Hello operation OperationID nexus-standalone-hello-op
2026/05/21 09:00:30 Hello result: Hello Temporal 👋
2026/05/21 09:00:30 ListNexusOperations results:
2026/05/21 09:00:30     OperationID: nexus-standalone-hello-op, Operation: say-hello, Status: Completed
2026/05/21 09:00:30     OperationID: nexus-standalone-echo-op, Operation: echo, Status: Completed
2026/05/21 09:00:30 Total Nexus operations: 2
```

If you run the starter code multiple times, you should see additional `ListNexusOperations` results, as more operations are run.
The same goes for the number from `CountNexusOperations`.
