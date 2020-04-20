## Exclusive-Choice Sample

This sample demonstrates how to run an activity based on a dynamic input.

### Steps to run this sample:
1) You need a Temporal service running. See details in README.md
2) Run the following command to start the worker
```
go run choice-exclusive/worker/main.go
```
3) Run the following command to start the exclusive choice workflow
```
go run choice-exclusive/starter/main.go
```
