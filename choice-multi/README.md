## Multi-Choice Sample

This sample demonstrates how to run multiple activities in parallel based on a dynamic input.

### Steps to run this sample:
1) Run a [Temporal service](https://github.com/temporalio/samples-go/tree/main/#how-to-use).
2) Run the following command to start the worker
```
go run choice-multi/worker/main.go
```
3) Run the following command to start the multi choice workflow
```
go run choice-multi/starter/main.go
```
