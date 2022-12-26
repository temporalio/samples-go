package capped_activities

import (
	"context"
	"math/rand"
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

// CappedActivitiesWorkflow workflow definition
func CappedActivitiesWorkflow(ctx workflow.Context, workerCount int) (result ChunkResult, err error) {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	selector := workflow.NewSelector(ctx)
	var totalItemCount, totalSum int

	batchSize := 30

	runActivity := func(i int) {
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

	// Start a batch of concurrent activities
	for i := 0; i < batchSize; i++ {
		runActivity(i)
	}

	// For the rest of the activities wait for one running
	// before adding a new one, this will keep the maximum
	// number of activities running in parallel to the size of the batch
	for j := batchSize; j < workerCount; j++ {
		// Each call to Select matches a single ready Future.
		// Each Future is matched only once independently on the number of Select calls.
		selector.Select(ctx)
		runActivity(j)

		if err != nil {
			return ChunkResult{}, err
		}
	}

	// Wait for the pending activities
	for i := 0; i < batchSize; i++ {
		selector.Select(ctx)
	}

	workflow.GetLogger(ctx).Info("Workflow completed.")

	return ChunkResult{totalItemCount, totalSum}, nil
}

func ChunkProcessingActivity(ctx context.Context, chunkID int) (result ChunkResult, err error) {
	// some fake processing logic here
	numberOfItemsInChunk := chunkID
	sumInChunk := chunkID * chunkID

	time.Sleep(time.Duration(1+rand.Intn(3)) * time.Second)

	activity.GetLogger(ctx).Info("Chunk processed", "chunkID", chunkID)
	return ChunkResult{numberOfItemsInChunk, sumInChunk}, nil
}
