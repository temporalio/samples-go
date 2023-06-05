This sample workflow demonstrates how to execute multiple activities in parallel and merge their results using futures.
The futures are awaited using Get method in the same order the activities are invoked. See `split-merge-selector` sample
to see how to process them in the order of activity completion instead.

### Steps to run this sample:

1) YRun a [Temporal service](https://github.com/temporalio/samples-go/tree/main/#how-to-use).
2) Run the following command to start the worker

```
go run splitmerge-future/worker/main.go
```

3) Run the following command to start the example

```
go run splitmerge-future/starter/main.go
```
