### Steps to run this sample:
1) Run a [Temporal service](https://github.com/temporalio/samples-go/tree/main/#how-to-use).

One way could be just to use the temporal CLI.  

```bash
temporal server start-dev --ui-port 8089
```

2) Run the following command to start the worker
```bash
go run otel/worker/main.go
```
3) Run the following command to start the example
```bash
go run otel/starter/main.go
```
