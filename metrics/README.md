### Steps to run this sample:
1) You need a Temporal service running. See details in README.md
2) Run the following command to start the worker
```
go run metrics/worker/main.go
```
3) Run the following command to start the example
```
go run metrics/starter/main.go
```
4) Check metrics at http://localhost:9090/metrics (this is where the Prometheus agent scrapes it from).
