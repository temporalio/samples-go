### Sleep for days

This sample demonstrates how to create a Temporal workflow that runs forever, sending an email every 30 days.

### Steps to run this sample:
1) Run a [Temporal service](https://github.com/temporalio/samples-go/tree/main/#how-to-use).
2) Run the following command to start the worker
```
go run worker/main.go
```
3) Run the following command to start the example
```
go run starter/main.go
```

This sample will run indefinitely until you send a `complete` signal to the workflow. See how to send a signal via Temporal CLI [here](https://docs.temporal.io/cli/workflow#signal).