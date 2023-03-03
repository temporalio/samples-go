### Update Sample

Here we show an example of a workflow representing a single integer counter
value that can be mutated via an update called `fetch_and_add` which adds its
argument to the counter and returns the original value of the counter. Negative
arguments will be rejected by the `fetch_and_add`'s associated validator and
thus will not be included in the workflow history.

### Steps to run this sample:
1) Run a [Temporal service](https://github.com/temporalio/samples-go/tree/main/#how-to-use).
2) Run the following command to start the worker
```
go run update/worker/main.go
```
3) Run the following command to start the example
```
go run update/starter/main.go
```
