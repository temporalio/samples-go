### Steps to run this sample:

1. Make sure you have a Temporal Cluster running or Temporal Cloud to connect to.
See the [Dev guide](https://docs.temporal.io/application-development/foundations#run-a-development-cluster) for the most up-to-date development option.

2. Start the Worker Process

```
go run yourapp/worker/main.go
```

3. Start the HTTP server

```
go run yourapp/gateway/main.go
```

4. Either in your browser, or via curl command hit `http://localhost:8001/start`
