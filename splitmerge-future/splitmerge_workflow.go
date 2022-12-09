package splitmerge_future

import (
	"context"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/workflow"
)

/**
 * This sample workflow demonstrates how to execute multiple activities in parallel and merge their results using futures.
 * The futures are awaited using Get method in the same order the activities are invoked. See `split-merge-selector` sample
 * to see how to process them in the order of activity completion instead.
 */

// ChunkResult contains the activity result for this sample
type ChunkResult struct {
	NumberOfItemsInChunk int
	SumInChunk           int
}

// SampleSplitMergeFutureWorkflow workflow definition
func SampleSplitMergeFutureWorkflow(ctx workflow.Context, processorCount int) (ChunkResult, error) {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var results []workflow.Future
	for i := 0; i < processorCount; i++ {
		// ExecuteActivity returns Future that doesn't need to be awaited immediately.
		future := workflow.ExecuteActivity(ctx, ChunkProcessingActivity, i+1)
		results = append(results, future)
	}

	var totalItemCount, totalSum int
	for i := 0; i < processorCount; i++ {
		var result ChunkResult
		// Blocks until the activity result is available.
		err := results[i].Get(ctx, &result)
		if err != nil {
			return ChunkResult{}, err
		}
		totalItemCount += result.NumberOfItemsInChunk
		totalSum += result.SumInChunk
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
