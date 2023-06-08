package batch_sliding_window

import (
	"fmt"
	"go.temporal.io/sdk/temporal"
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

	if input.SlidingWindowSize < input.Partitions {
		return 0, temporal.NewApplicationError(
			"SlidingWindowSize cannot be less than number of partitions", "invalidInput")
	}
	partitions := divideIntoPartitions(recordCount, input.Partitions)
	windowSizes := divideIntoPartitions(input.SlidingWindowSize, input.Partitions)

	workflow.GetLogger(ctx).Info("ProcessBatchWorkflow",
		"input", input,
		"recordCount", recordCount,
		"partitions", partitions,
		"windowSizes", windowSizes)

	var results []workflow.ChildWorkflowFuture
	offset := 0
	for i := 0; i < input.Partitions; i++ {
		// Makes child id more user-friendly
		childId := fmt.Sprintf("%s/%d", workflow.GetInfo(ctx).WorkflowExecution.ID, i)
		childCtx := workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{WorkflowID: childId})
		// Define partition boundaries.
		maximumPartitionOffset := offset + partitions[i]
		if maximumPartitionOffset > recordCount {
			maximumPartitionOffset = recordCount
		}
		input := SlidingWindowWorkflowInput{
			PageSize:          input.PageSize,
			SlidingWindowSize: windowSizes[i],
			Offset:            offset,                 // inclusive
			MaximumOffset:     maximumPartitionOffset, // exclusive
		}
		child := workflow.ExecuteChildWorkflow(childCtx, SlidingWindowWorkflow, input)
		results = append(results, child)
		offset += partitions[i]
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

func divideIntoPartitions(number int, n int) []int {
	base := number / n
	remainder := number % n
	partitions := make([]int, n)

	for i := 0; i < n; i++ {
		partitions[i] = base
	}

	for i := 0; i < remainder; i++ {
		partitions[i] += 1
	}

	return partitions
}
