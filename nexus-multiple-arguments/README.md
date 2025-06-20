# Nexus Context Propagation

This sample shows how to map a Nexus operation to a caller workflow that takes multiple input arguments using [temporalnexus.NewWorkflowRunOperationWithOptions](https://pkg.go.dev/go.temporal.io/sdk/temporalnexus#MustNewWorkflowRunOperationWithOptions)

For more details on Nexus and how to set up to run this sample, please see the [Nexus Sample](../nexus/README.md).

### Running the sample

In separate terminal windows:

### Nexus handler worker

```
cd handler
go run ./worker \
    -target-host localhost:7233 \
    -namespace my-target-namespace
```

### Nexus caller worker

```
cd caller
go run ./worker \
    -target-host localhost:7233 \
    -namespace my-caller-namespace
```

### Start caller workflow

```
cd caller
go run ./starter \
    -target-host localhost:7233 \
    -namespace my-caller-namespace
```

### Output

which should result in:
```
2025/02/27 12:57:40 Started workflow WorkflowID nexus_hello_caller_workflow_20240723195740 RunID c9789128-2fcd-4083-829d-95e43279f6d7
2025/02/27 12:57:40 Workflow result: ¡Hola! Nexus, caller-id: samples-go 👋
```
