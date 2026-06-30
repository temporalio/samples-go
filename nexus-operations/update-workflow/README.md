# Nexus Operations: UpdateWorkflow 

This sample shows how to expose and run an `UpdateWorkflow` Temporal operation as a Nexus operation. The operation sends an update to a running workflow - so it is required that the workflowID be known and that the worklow isn't yet in a terminal state just as is the case for a normal `UpdateWorkflow` operation

For Nexus Operation wrapped `UpdateWorkflow` operations, only `WorkflowUpdateStageAccepted` is accepted as the `WaitForStage`. This is because nexus ops need to complete within 10s and `WorkflowUpdateStageCompleted` will wait until the update is fully finished - this could lead to cases where Updates keep getting retried on a consistently slow operation

## Run the samples

### Create caller and target namespaces

```
temporal operator namespace create --namespace my-target-namespace
temporal operator namespace create --namespace my-caller-namespace
```

### Create the Nexus endpoint

```
temporal operator nexus endpoint create \
  --name counter-update-endpoint \
  --target-namespace my-target-namespace \
  --target-task-queue counter-update-handler-tq
```

### Setting up the workers 

#### Handler worker (target namespace)

```
go run ./nexus-operations/update-workflow/handler/worker -namespace my-target-namespace
```

#### Caller worker (caller namespace)

```
go run ./nexus-operations/update-workflow/caller/worker -namespace my-caller-namespace
```

### Set up the receiver 

Start the target workflow (target namespace) so that it can receive the updates 

```
go run ./nexus-operations/update-workflow/handler/starter -namespace my-target-namespace
```

Send a `done` signal to the counter workflow to close it  

```
temporal workflow signal --namespace my-target-namespace --workflow-id counter-workflow-1 --name done
```

### Trigger the Nexus UpdateWorkflow operation

```
go run ./nexus-operations/update-workflow/caller/starter -namespace my-caller-namespace 
```


