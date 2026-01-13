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
go run helloworld/worker/main.go
```

You should see two console log lines:
 1. Creating the logger 
 2. Starting the Worker with Namespace `default`, and TaskQueue `hello-world` and it will list the WorkerID for the created worker.  

For example:
```bash
2025/12/22 15:00:15 INFO  No logger configured for temporal client. Created default one.

2025/12/22 15:00:16 INFO  Started Worker Namespace default TaskQueue hello-world WorkerID 82087
```

> [!NOTE]
> Timestamps and IDs will differ on your machine.

3) In a third terminal, run the following command to start the example
```bash
go run helloworld/starter/main.go
```

You should see three console log lines: 1) Creating the logger, 2) Starting the workflow, and 3) The workflow result. 

For example:
```bash
2025/12/22 15:07:24 INFO  No logger configured for temporal client. Created default one. 

2025/12/22 15:07:25 Started workflow WorkflowID hello_world_workflowID RunID 019b47ac-7c4d-701f-9d35-3acfe171723e

2025/12/22 15:07:25 Workflow result: Hello Temporal!
```

## Troubleshooting

> [!NOTE]
> This sample relies on the existence of the `default` namespace. If you are a Temporal contributor and built the server locally using [these instructions](https://github.com/temporalio/temporal/blob/main/CONTRIBUTING.md) instead of the dev server in step 1, the `default` namespace is not created automatically. You will need to create it using the instructions below or you will get an error message when starting the worker: `Unable to start worker Namespace default is not found.`
>
>Confirm that the `default` namespace exists using this command: 
>
>```bash
>temporal operator namespace list 
>```
>You should see a line with `NamespaceInfo.Name` set to `default`.
>
>If you do not find a namespace named `default`, you can create it here using this command:
>
>```bash
>temporal operator namespace create -n default 
>```
