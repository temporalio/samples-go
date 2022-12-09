package splitmerge_selector

import (
	"context"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/workflow"
)

/**
 * This sample workflow demonstrates how to execute multiple activities in parallel and merge their results using futures.
 * The futures are awaited using Selector. It allows processing them as soon as they become ready. See `split-merge-future` sample
 * to see how to process them without Selector in the order of activity invocation instead.
 */

// ChunkResult contains the activity result for this sample
type ChunkResult struct {
	NumberOfItemsInChunk int
	SumInChunk           int
}

// SampleSplitMergeSelectorWorkflow workflow definition
func SampleSplitMergeSelectorWorkflow(ctx workflow.Context, workerCount int) (result ChunkResult, err error) {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	selector := workflow.NewSelector(ctx)
	var totalItemCount, totalSum int
	for i := 0; i < workerCount; i++ {
		// ExecuteActivity returns Future that doesn't need to be awaited immediately.
		future := workflow.ExecuteActivity(ctx, ChunkProcessingActivity, i+1)
		selector.AddFuture(future, func(f workflow.Future) {
			var r ChunkResult
			err1 := f.Get(ctx, &r)
			if err1 != nil {
				err = err1
				return
			}
			totalItemCount += r.NumberOfItemsInChunk
			totalSum += r.SumInChunk
		})
	}

	for i := 0; i < workerCount; i++ {
		// Each call to Select matches a single ready Future.
		// Each Future is matched only once independently on the number of Select calls.
		selector.Select(ctx)
		if err != nil {
			return ChunkResult{}, err
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
