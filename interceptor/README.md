# Interceptor Sample

This sample shows how to make a worker interceptor that intercepts workflow and activity `GetLogger` calls to customize
the logger.

### Steps to run this sample:
1) You need a Temporal service running. See details in README.md
2) Run the following command to start the worker
```
go run ./interceptor/worker
```
3) Run the following command to start the example
```
go run ./interceptor/starter
```

Notice the log output has the `WorkflowStartTime`/`ActivityStartTime` tags on the logs.
