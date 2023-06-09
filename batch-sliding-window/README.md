## Sliding Window Batch Sample 

A sample implementation of a batch processing Workflow that maintains a sliding window of record processing Workflows.

A SlidingWindowWorkflow starts a configured number (sliding window size) of RecordProcessorWorkflow children in parallel. 
Each child processes a single record. When a child completes a new child is started.

A SlidingWindowWorkflow calls continue-as-new after starting a preconfigured number of children to keep its history size bounded.
A RecordProcessorWorkflow reports its completion through a Signal to its parent.
This allows to notify a parent that called continue-as-new.

A single instance of SlidingWindowWorkflow has limited window size and throughput. 
To support larger window size and overall throughput multiple instances of SlidingWindowWorkflow run in parallel.

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
