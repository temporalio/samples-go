package splitmerge

import (
	"context"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/workflow"
)

/**
* This sample workflow demonstrates how to use multiple Temporal corotinues (instead of native goroutine) to process a
* chunk of a large work item in parallel, and then merge the intermediate result to generate the final result.
* In Temporal workflow, you should not use go routine. Instead, you use corotinue via workflow.Go method.
 */

type (
	// ChunkResult contains the result for this sample
	ChunkResult struct {
		NumberOfItemsInChunk int
		SumInChunk           int
	}
)

// SampleSplitMergeWorkflow workflow definition
func SampleSplitMergeWorkflow(ctx workflow.Context, workerCount int) (ChunkResult, error) {
	chunkResultChannel := workflow.NewChannel(ctx)
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	for i := 1; i <= workerCount; i++ {
		chunkID := i
		workflow.Go(ctx, func(ctx workflow.Context) {
			var result ChunkResult
			err := workflow.ExecuteActivity(ctx, ChunkProcessingActivity, chunkID).Get(ctx, &result)
			if err == nil {
				chunkResultChannel.Send(ctx, result)
			} else {
				chunkResultChannel.Send(ctx, err)
			}
		})
	}

	var totalItemCount, totalSum int
	for i := 1; i <= workerCount; i++ {
		var v interface{}
		chunkResultChannel.Receive(ctx, &v)
		switch r := v.(type) {
		case error:
		// failed to process this chunk
		// some proper error handling code here
		case ChunkResult:
			totalItemCount += r.NumberOfItemsInChunk
			totalSum += r.SumInChunk
		}
	}

	workflow.GetLogger(ctx).Info("Workflow completed.")

	return ChunkResult{totalItemCount, totalSum}, nil
}

func ChunkProcessingActivity(ctx context.Context, chunkID int) (result ChunkResult, err error) {
	// some fake processing logic here
	numberOfItemsInChunk := chunkID
	sumInChunk := chunkID * chunkID

	activity.GetLogger(ctx).Info("Chunk processed", "chunkID", chunkID)
	return ChunkResult{numberOfItemsInChunk, sumInChunk}, nil
}
