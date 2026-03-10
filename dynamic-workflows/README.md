# Dynamic Workflows and Activities

The purpose of this sample is to demonstrate registering and using dynamic workflows and activities to
handle various workflow types at runtime.

### Steps to run this sample:
1) Run a [Temporal service](https://github.com/temporalio/samples-go/tree/main/#how-to-use).
2) Run the following command to start the worker
```
go run dynamic-workflows/worker/main.go
```
3) Run the following command to start the example
```
go run dynamic-workflows/starter/main.go
```
