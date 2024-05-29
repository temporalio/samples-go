* The sample demonstrates how to deal with multiple signals that can come out of order and require actions
* if a certain signal not received in a specified time interval.

This specific sample receives three signals: Signal1, Signal2, Signal3. They have to be processed in the
sequential order, but they can be received out of order.
There are two timeouts to enforce.
The first one is the maximum time between signals.
The second limits the total time since the first signal received.

A naive implementation of such use case would use a single loop that contains a Selector to listen on three
signals and a timer. Something like:

	for {
		selector := workflow.NewSelector(ctx)
		selector.AddReceive(workflow.GetSignalChannel(ctx, "Signal1"), func(c workflow.ReceiveChannel, more bool) {
			// Process signal1
		})
		selector.AddReceive(workflow.GetSignalChannel(ctx, "Signal2"), func(c workflow.ReceiveChannel, more bool) {
			// Process signal2
		}
		selector.AddReceive(workflow.GetSignalChannel(ctx, "Signal3"), func(c workflow.ReceiveChannel, more bool) {
			// Process signal3
		}
		cCtx, cancel := workflow.WithCancel(ctx)
		timer := workflow.NewTimer(cCtx, timeToNextSignal)
		selector.AddFuture(timer, func(f workflow.Future) {
			// Process timeout
		})
 		selector.Select(ctx)
	    cancel()
      // break out of the loop on certain condition
	}

The above implementation works. But it quickly becomes pretty convoluted if the number of signals
and rules around order of their arrivals and timeouts increases.

The following example demonstrates an alternative approach. It receives signals in a separate goroutine.
Each signal handler just updates a correspondent shared variable with the signal data.
The main workflow function awaits the next step using `workflow.AwaitWithTimeout` using condition composed of
the shared variables. This makes the main workflow method free from signal callbacks and makes the business logic
clear.

### Steps to run this sample:

1) You need a Temporal service running. See details in README.md
2) Run the following command to start the worker

```
go run await-signals/worker/main.go
```

3) Run the following command to start the workflow and send signals in random order

```
go run await-signals/starter/main.go
```
