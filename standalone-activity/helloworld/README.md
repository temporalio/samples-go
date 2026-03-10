This sample demonstrates how to use a Standalone Activity (executing an Activity without wrapping it in a Workflow).
It also shows you how to use the `ListActivities` and `CountActivities` APIs.

## NOTE: This new feature is not ready for use yet. It will only work once we release a special CLI server for pre-release, once that happens, this README will be updated.

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

You should see something similar to the following output.

For example:
```bash
2026/02/23 14:12:00 INFO  No logger configured for temporal client. Created default one.
2026/02/23 14:12:00 Started standalone activity ActivityID standalone_activity_helloworld_ActivityID RunID 019c8c8f-324f-7c06-a92e-a9f7e612ce69
2026/02/23 14:12:00 Activity result: Hello Temporal!
2026/02/23 14:12:00 ListActivity results
2026/02/23 14:12:00 	ActivityID: standalone_activity_helloworld_ActivityID, Type: Activity, Status: Completed
2026/02/23 14:12:00 CountActivities: 1
```

If you run the starter code multiple times, you should see additional `ListActivity` results, as more activites are run.
The same goes for the number of activities from `CountActivities`.
