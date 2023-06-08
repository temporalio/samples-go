## Sliding Window Batch Sample 

A sample implementation of a batch processing Workflow that maintains a sliding window of record processing Workflows.

A Workflow starts a configured number of Child Workflows in parallel. Each child processes a single record.
When a child completes a new child immediately started.

A Parent Workflow calls continue-as-new after starting a preconfigured number of children.
A child completion is reported through a Signal as a parent cannot directly wait for a child started by a previous run.

Multiple instances of SlidingWindowBatchWorkflow run in parallel each processing a subset of records to support higher total rate of processing.

#### Running the Sliding Window Batch Sample

Make sure the [Temporal Server is running locally](https://docs.temporal.io/application-development/foundations#run-a-development-cluster).

From the root of the project, start a Worker:

```bash
go run batch-sliding-window/worker/main.go
```

Start the Workflow Execution:

```bash
go run batch-sliding-window/starter/main.go
```
