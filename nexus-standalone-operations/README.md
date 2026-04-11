This sample demonstrates how to use Standalone Nexus Operations (executing Nexus operations directly from client code without wrapping them in a Workflow).
It shows both sync and async (workflow-backed) operations, and how to use the `ListNexusOperations` and `CountNexusOperations` APIs.

## Note: Standalone Nexus operations require a server version that supports this feature.

### Steps to run this sample (with expected output):
1) Run a [Temporal server](https://github.com/temporalio/samples-go/tree/main/#how-to-use). (If you are going to run locally, you will want to start it in another terminal; this command is blocking and runs until it receives a SIGINT (Ctrl + C) command.)

If you used the above instructions to start the server, you should see a line about the CLI, Server and UI versions, and one line each for the Server URL, UI URL and Metrics endpoint. It should look something like this:

```bash
> temporal server start-dev
CLI 1.5.1 (Server 1.29.1, UI 2.42.1)

Server:  localhost:7233
UI:      http://localhost:8233
Metrics: http://localhost:57058/metrics
```

2) Create a Nexus endpoint that routes to the worker's task queue. In a second terminal, run:
```bash
temporal operator nexus endpoint create \
  --name nexus-standalone-operations-endpoint \
  --target-namespace default \
  --target-task-queue nexus-standalone-operations
```

3) Open a third terminal, and run the following command to start the worker. The worker is a blocking process that runs until it receives a SIGINT (Ctrl + C) command.
```bash
go run nexus-standalone-operations/worker/main.go
```

You should see two console log lines:
1. Creating the logger
2. Starting the Worker with Namespace `default`, and TaskQueue `nexus-standalone-operations` and it will list the WorkerID for the created worker.

For example:
```bash
2026/04/11 15:00:15 INFO  No logger configured for temporal client. Created default one.
2026/04/11 15:00:16 INFO  Started Worker Namespace default TaskQueue nexus-standalone-operations WorkerID 82087
```

> [!NOTE]
> Timestamps and IDs will differ on your machine.

4) In a fourth terminal, run the following command to start the example:
```bash
go run nexus-standalone-operations/starter/main.go
```

You should see something similar to the following output:

```bash
2026/04/11 14:12:00 INFO  No logger configured for temporal client. Created default one.
2026/04/11 14:12:00 Started Echo operation OperationID nexus-standalone-echo-op
2026/04/11 14:12:00 Echo result: hello
2026/04/11 14:12:00 Started Hello operation OperationID nexus-standalone-hello-op
2026/04/11 14:12:00 Hello result: Hello Temporal 👋
2026/04/11 14:12:00 ListNexusOperations results:
2026/04/11 14:12:00 	OperationID: nexus-standalone-echo-op, Operation: echo, Status: Completed
2026/04/11 14:12:00 	OperationID: nexus-standalone-hello-op, Operation: say-hello, Status: Completed
2026/04/11 14:12:00 Total Nexus operations: 2
```

If you run the starter code multiple times, you should see additional `ListNexusOperations` results, as more operations are run.
The same goes for the number from `CountNexusOperations`.
