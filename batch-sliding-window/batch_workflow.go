package batch_sliding_window

import (
	"fmt"
	"go.temporal.io/sdk/workflow"
	"time"
)

// ProcessBatchWorkflowInput input of the ProcessBatchWorkflow.
// A single input structure is preferred to multiple workflow arguments to simplify backward compatible API changes.
type ProcessBatchWorkflowInput struct {
	PageSize          int // Number of children started by a single sliding window workflow run
	SlidingWindowSize int // Maximum number of children to run in parallel.
	Partitions        int // How many sliding windows to run in parallel.
}

// ProcessBatchWorkflow sample Partitions the data set into continuous ranges.
// A real application can choose any other way to divide the records into multiple collections.
func ProcessBatchWorkflow(ctx workflow.Context, input ProcessBatchWorkflowInput) (processed int, err error) {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Second,
	})

	var recordLoader *RecordLoader // RecordLoader activity reference
	var recordCount int
	err = workflow.ExecuteActivity(ctx, recordLoader.GetRecordCount).Get(ctx, &recordCount)
	if err != nil {
		return 0, err
	}

	partitionSize := recordCount / input.Partitions
	if recordCount%input.Partitions > 0 {
		partitionSize += 1
	}

	// Divide the window size between partitions
	partitionWindowSize := input.SlidingWindowSize / input.Partitions
	lastPartitionWindowSize := input.SlidingWindowSize % input.Partitions
	if lastPartitionWindowSize == 0 {
		lastPartitionWindowSize = partitionWindowSize
	}

	var results []workflow.ChildWorkflowFuture
	for i := 0; i < input.Partitions; i++ {
		// Makes child id more user-friendly
		childId := fmt.Sprintf("%s/%d", workflow.GetInfo(ctx).WorkflowExecution.ID, i)
		childCtx := workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{WorkflowID: childId})
		// Define partition boundaries.
		offset := partitionSize * i             // inclusive
		maximumOffset := offset + partitionSize // exclusive
		if maximumOffset > recordCount {
			maximumOffset = recordCount
		}
		windowSize := partitionWindowSize
		if i == input.Partitions-1 {
			windowSize = lastPartitionWindowSize
		}
		input := SlidingWindowWorkflowInput{
			PageSize:          input.PageSize,
			SlidingWindowSize: windowSize,
			Offset:            offset,
			MaximumOffset:     maximumOffset,
		}
		child := workflow.ExecuteChildWorkflow(childCtx, SlidingWindowWorkflow, input)
		results = append(results, child)
	}
	// Waits for all child workflows to complete
	result := 0
	for _, partitionResult := range results {
		var r int
		err := partitionResult.Get(ctx, &r) // blocks until the child completion
		if err != nil {
			return 0, err
		}
		result += r
	}
	return result, nil
}
