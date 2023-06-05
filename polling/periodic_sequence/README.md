## Periodic sequence

This samples shows periodic polling via Child Workflow.

This is a rare scenario where polling requires execution of a sequence of Activities, or Activity arguments need to change between polling retries.

For this case we use a Child Workflow to call polling Activities a set number of times in a loop and then periodically calls continue-as-new.

The Parent Workflow is not aware about the Child Workflow calling continue-as-new and it gets notified when it completes (or fails).

### Steps to run this sample:
1) You need a Temporal service running. See details in README.md
2) Run the following command to start the worker
```
go run periodic_sequence/worker/main.go
```
3) Run the following command to start the example
```
go run periodic_sequence/starter/main.go
```
