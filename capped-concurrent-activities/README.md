This sample workflow demonstrates how to execute multiple activities in parallel running just a maximum at the time per workflow.

### Steps to run this sample:

1. You need a Temporal service running. See details in README.md
2. Run the following command to start the worker

```
go run capped-concurrent-activities/worker/main.go
```

3. Run the following command to start the example

```
go run capped-concurrent-activities/starter/main.go
```

