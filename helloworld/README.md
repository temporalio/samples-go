### Steps to run this sample (with expected output):
1) Run a [Temporal server](https://github.com/temporalio/samples-go/tree/main/#how-to-use). (If you are going to run locally, you will want to open and run in another terminal as the command is a blocking process that runs until it receives a SIGINT (Ctrl + C) command.)


You should see initialization logs, the last will will indicate `workflow successfully started`:
```
{"level":"info","ts":"2025-12-22T14:53:19.673-0500","msg":"workflow successfully started","service":"worker","wf-type":"temporal-sys-history-scanner-workflow","logging-call-at":"/repo/temporal/service/worker/scanner/scanner.go:273"}
```

2) Create a new namespace named `default` using this command: 
```bash
temporal operator namespace create -n default 
```

To confirm the namespace was created successfully you can run:

```bash
temporal operator namespace list 
```

You should see a line with `NamespaceInfo.Name` set to `default`.

3) Open another terminal, and run the following command to start the worker. The worker is a blocking process that runs until it receives a SIGINT (Ctrl + C) command.)
```
go run helloworld/worker/main.go
```

You should see two console log lines about 1) Creating the logger and 2) Starting the Worker with Namespace `default`, and TaskQueue `hello-world` and it will list the  WorkerID for the created worker.  

For example:
```
2025/12/22 15:00:15 INFO  No logger configured for temporal client. Created default one.

2025/12/22 15:00:16 INFO  Started Worker Namespace default TaskQueue hello-world WorkerID 82087
```

4) From your terminal window, run the following command to start the example
```
go run helloworld/starter/main.go
```

You should see three console log lines about 1) Creating the logger, 2) Starting the workflow, and 3) The workflow result. 

For example:
```
2025/12/22 15:07:24 INFO  No logger configured for temporal client. Created default one. 

2025/12/22 15:07:25 Started workflow WorkflowID hello_world_workflowID RunID 019b47ac-7c4d-701f-9d35-3acfe171723e

2025/12/22 15:07:25 Workflow result: Hello Temporal!
```