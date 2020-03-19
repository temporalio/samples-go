This sample workflow demos a file processing process. The key part is to show how to use the session API.

The workflow first starts an activity to download a requested resource file from web and store it locally on the host where it runs the download activity. Then, the workflow will start more activities to process the downloaded resource file. The key part is the following activities have to be run on the same host as the initial downloading activity. This is achieved by using the session API.

Steps for using Session API:
1) When starting worker, set `EnableSessionWorker` to true in workerOptions.
2) In the workflow code, create a new session using the `CreateSession()` API
```
  so := &workflow.SessionOptions{
    CreationTimeout:  time.Minute,
    ExecutionTimeout: time.Minute,
  }
  sessionCtx, err := workflow.CreateSession(ctx, so)
```
3) Use the returned `sessionCtx` or its child context to execute activities. These activities will be to scheduled on the same host.
4) After all activities are executed, call `CompleteSession()`.
```
  workflow.CompleteSession(sessionCtx)
```
5) Check the inline document in workflow/session.go of the Go SDK repo for more advanced usage.

Steps to run this sample:
1) You need a Temporal service running. See details in README.md
2) Run the following command multiple times on different console window. This is to simulate running workers on multiple different machines.
```
go run fileprocessing/worker/main.go
```
3) Run the following command to submit a start request for this fileprocessing workflow.
```
go run fileprocessing/starter/main.go
```

You should see that all activities for one particular workflow execution are scheduled to run on one console window.
