# Nexus Context Propagation

This sample shows how to propagate context through client calls, workflows, and Nexus headers.

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
2025/02/28 14:54:29 Started workflow WorkflowID nexus_hello_caller_workflow_20250228145225 RunID 01954ec2-d1d2-72b7-a6cc-a83de5421754
2025/02/28 14:54:30 Workflow result: Nexus Echo 👋, caller-id: samples-go
2025/02/28 14:54:30 Started workflow WorkflowID nexus_hello_caller_workflow_20250228145430 RunID 01954ec4-bc4d-7dac-a311-d4ceacacfa9a
2025/02/28 14:54:31 Workflow result: ¡Hola! Nexus, caller-id: samples-go 👋
```
