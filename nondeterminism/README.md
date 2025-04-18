# Non-deterministic error examples

These samples showcase various cases in which you can run into non-deterministic errors. The steps explaining how to run the code such that such error is thrown are explained in each workflow sample.

See [the deterministic constraints](https://docs.temporal.io/workflow-definition#deterministic-constraints) as a docs reference.

### Steps to run this sample:
1) Run a [Temporal service](https://github.com/temporalio/samples-go/tree/main/#how-to-use).
2) Run the following command to start the worker
```
go run nondeterminism/worker/main.go
```
3) Run the following command to start the example
```
go run nondeterminism/starter/main.go
```
4) Check the instruction for each individual workflow to trigger the NDE
