# Invoking Activities by Name

The purpose of this sample is to demonstrate invocation of workflows and activities using name 
rather than strongly typed function.

### Steps to run this sample:
1) You need a Temporal service running. See details in README.md
2) Run the following command to start the worker
```
go run dynamic/worker/main.go
```
3) Run the following command to start the example
```
go run dynamic/starter/main.go
```
