This sample demonstrates getting multiple workflow histories and replaying them.

### Steps to run this sample:
1) Run a [Temporal service](https://github.com/temporalio/samples-go/tree/main/#how-to-use).
2) Run the following command to start the worker
```shell script
go run multi-history-replay/worker/main.go
```
3) Run the following command to start mutiple parallel workflows. This simulates having existing workflows
on the server.
```shell script
go run multi-history-replay/starter/main.go
```
4) Run the following command to replay the histories generated in step 3
```shell script
go run multi-history-replay/replayer/main.go
```
