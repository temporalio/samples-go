# yourapp Temporal Go SDK sample

This application provides a basic foundation for writing Temporal Appplications using the Go SDK.

The sample is intended to contain a great deal of information about Temporal code structure within the code comments.

# How to run

1. Make sure you have a Temporal Cluster running or Temporal Cloud to connect to.
See the [Dev guide](https://docs.temporal.io/application-development/foundations#run-a-development-cluster) for the most up-to-date development option.

2. Start the Worker Process

```
go run yourapp/worker/main_dacx.go
```

3. Start the HTTP server

```
go run yourapp/gateway/main_dacx.go
```

4. Either in your browser, or via curl command hit `http://localhost:8091/start`
