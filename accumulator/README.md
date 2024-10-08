# Accumulator
This sample demonstrates how to accumulate many signals (events) over a time period. 
This sample implements the Accumulator Pattern: collect many meaningful things that need to be collected and worked on together, such as all payments for an account, or all account updates by account.
 
This sample models robots being created throughout the time period, groups them by what color they are, and greets all the robots of a color at the end.
 
A new workflow is created per grouping.
A sample activity at the end is given, and you could add an activity to
process individual events in the processGreeting() method.

Because Temporal Workflows cannot have an unlimited size, Continue As New is used to process more signals that may come in.
You could create as many groupings as desired, as Temporal Workflows scale out to many workflows without limit.

You could vary the time that the workflow waits for other signals, say for a day, or a variable time from first signal with the GetNextTimeout() function.
This example supports exiting early with an exit signal. Pending greetings are still collected after exit signal is sent.


### Steps to run this sample:

1) You need a Temporal service running. See details in repo's README.md
2) Run the following command to start the worker

```
go run accumulator/worker/main.go
```

3) Run the following command to start the workflow and send signals in random order

```
go run accumulator/starter/main.go
```

You can also run tests with
```
go test accumulator/accumulate_signals_workflow_test.go
```
