This sample workflow demonstrates how to execute multiple activities in parallel and merge their results using futures.
The futures are awaited using Selector. It allows processing them as soon as they become ready. See `split-merge-future`
sample to see how to process them without Selector in the order of activity invocation instead.

### Steps to run this sample:

1) Run a [Temporal service](https://github.com/temporalio/samples-go/tree/main/#how-to-use).
2) Run the following command to start the worker

```
go run splitmerge-selector/worker/main.go
```

3) Run the following command to start the example

```
go run splitmerge-selector/starter/main.go
```
