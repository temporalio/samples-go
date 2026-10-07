This sample shows how to implement a workflow that simulates a queue, using the event history and workflow input. It continuously listens for incoming tasks via a channel, while also listening on a different channel to understand when to write a batch. The tasks are sent from a different workflow through Temporal signals and a repeating timer writes on the other channel to signal the workflow when to write a batch.

The workflow also uses continue-as-new to avoid large history size caused by the timers and signals. This is done when a certain number of events (signal received, timer fired) have been recorded.

The example also illustrates how to simulate a ticker using timers.

### Steps to run this sample:
1) Run a [Temporal service](https://github.com/temporalio/samples-go/tree/main/#how-to-use).
2) Run the following command to start the worker
```
go run forever-batch-operations/worker/main.go
```
3) Run the following command to start the example
```
go run forever-batch-operations/starter/main.go
```

Note that the workflows will continue running even after you stop the worker.

Compare the values_received.txt and values_sent.txt files. Values should be written in the same order.
