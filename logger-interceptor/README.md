# Interceptor Sample

This sample shows how to make a worker interceptor that intercepts workflow and activity `GetLogger` calls to customize
the logger.

### Steps to run this sample:
1) Run a [Temporal service](https://github.com/temporalio/samples-go/tree/main/#how-to-use).
2) Run the following command to start the worker
```
go run ./logger-interceptor/worker
```
3) Run the following command to start the example
```
go run ./logger-interceptor/starter
```

Notice the log output has the `WorkflowStartTime`/`ActivityStartTime` tags on the logs.
