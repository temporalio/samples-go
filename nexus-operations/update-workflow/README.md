# Nexus Operations: UpdateWorkflow 

This sample shows how to expose and run an `UpdateWorkflow` Temporal operation as a Nexus operation. The operation sends an update to a running workflow - so it is required that the workflowID be known and that the worklow isn't yet in a terminal state just as is the case for a normal `UpdateWorkflow` operation

For Nexus Operation wrapped `UpdateWorkflow` operations, only `WorkflowUpdateStageAccepted` is accepted as the `WaitForStage`. This is because nexus ops need to complete within 10s and `WorkflowUpdateStageCompleted` will wait until the update is fully finished - this could lead to cases where Updates keep getting retried on a consistently slow operation

## Running the samples

There are two samples in this repo 
1. Nexus Op UpdateWorkfow inside in a workflow. This is for cases where the Nexus Op will be called inside a workflow - see `caller` folder for details
2. Standalone Nexus Op Update Workflow. This is for use-cases where the Nexus Op needs to be called from outside of a workflow. See the `standalone` folder for details 

For both cases, the handler remains the same 

### Running Standalone Nexus Operation UpdateWorkflow

Ensure the temporal server has the required configurations. If required, start the dev server locally with the dynamic config flags required for standalone Nexus operations:

TODO: add/estimate version of the temporal server that has the SANO feature(behind gates, etc)

```
temporal server start-dev \
  --dynamic-config-value "nexusoperation.enableStandalone=true" \
  --dynamic-config-value "history.enableCHASMCallbacks=true" \
  --dynamic-config-value "history.enableUpdateCallbacks=true"
```

#### Create the handler namespace

```
temporal operator namespace create --namespace my-target-namespace
```

#### Create the Nexus endpoint

```
temporal operator nexus endpoint create \
  --name counter-update-endpoint \
  --target-namespace my-target-namespace \
  --target-task-queue counter-update-handler-tq
```

#### Start the Handler worker 

```
go run ./nexus-operations/update-workflow/handler/worker -namespace my-target-namespace
```

#### Start the Handler workflow (receiver)

Start the handler workflow so that it can receive the updates 

```
go run ./nexus-operations/update-workflow/handler/starter -namespace my-target-namespace
```

NOTE:
Send a `done` signal to the counter workflow to close it  

```
temporal workflow signal --namespace my-target-namespace --workflow-id counter-workflow-1 --name done
```

#### Run the Standalone Nexus Operation 

```
go run ./nexus-operations/update-workflow/standalone -namespace my-target-namespace -incr 1
```

### Running Nexus Operation UpdateWorkflow inside a Workflow

#### Create the caller and handler namespaces

```
temporal operator namespace create --namespace my-target-namespace
temporal operator namespace create --namespace my-caller-namespace
```

#### Create the Nexus Endpoint

```
temporal operator nexus endpoint create \
  --name counter-update-endpoint \
  --target-namespace my-target-namespace \
  --target-task-queue counter-update-handler-tq
```

#### Setting up the Workers 

##### Handler worker (target namespace)

```
go run ./nexus-operations/update-workflow/handler/worker -namespace my-target-namespace
```

##### Caller worker (caller namespace)

```
go run ./nexus-operations/update-workflow/caller/worker -namespace my-caller-namespace
```

#### Set up the receiver 

Start the handler workflow so that it can receive the updates 

```
go run ./nexus-operations/update-workflow/handler/starter -namespace my-target-namespace
```

NOTE:
Send a `done` signal to the counter workflow to close it  

```
temporal workflow signal --namespace my-target-namespace --workflow-id counter-workflow-1 --name done
```

#### Trigger the Nexus UpdateWorkflow operation

```
go run ./nexus-operations/update-workflow/caller/starter -namespace my-caller-namespace 
```


