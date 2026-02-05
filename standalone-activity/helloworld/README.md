This sample demonstrates how to use a Standalone Activity (executing an Activity without wrapping it in a Workflow)

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

2) Open a second terminal, and run the following command to start the worker. The worker is a blocking process that runs until it receives a SIGINT (Ctrl + C) command.
```bash
go run standalone-activity/helloworld/worker/main.go
```

You should see two console log lines:
1. Creating the logger
2. Starting the Worker with Namespace `default`, and TaskQueue `standalone-activity-helloworld` and it will list the WorkerID for the created worker.

For example:
```bash
2025/12/22 15:00:15 INFO  No logger configured for temporal client. Created default one.

2025/12/22 15:00:16 INFO  Started Worker Namespace default TaskQueue standalone-activity-helloworld WorkerID 82087
```

> [!NOTE]
> Timestamps and IDs will differ on your machine.

3) In a third terminal, run the following command to start the example
```bash
go run standalone-activity/helloworld/starter/main.go
```

You should see two console log lines: 1) Creating the logger, 2) The standalone activity result

For example:
```bash
2026/02/05 11:30:47 INFO  No logger configured for temporal client. Created default one.

2026/02/05 11:30:47 Started standalone activity ActivityID standalone_activity_helloworld_ActivityID RunID 019c2f49-1ff1-7a44-beee-7ff4b36ecc27

2026/02/05 11:30:47 Activity result: Hello Temporal!
```
