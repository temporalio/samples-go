### Update Cancel Sample

Here we show an example of a workflow with a long running update. Through the use of an interceptor we are able to cancel the update by sending another special "cancel" update.

### Steps to run this sample:
1) Run a [Temporal service](https://github.com/temporalio/samples-go/tree/main/#how-to-use).
2) Run the following command to start the worker
```
go run update-cancel/worker/main.go
```
3) Run the following command to start the example
```
go run update-cancel/starter/main.go
```
