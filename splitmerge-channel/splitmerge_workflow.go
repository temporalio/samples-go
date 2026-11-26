package splitmerge_channel

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

type ActivityNResult struct {
	ID int
}
type ActivityMResult struct {
}

// SampleSplitMergeChannelWorkflow workflow definition
func SampleSplitMergeChannelWorkflow(ctx workflow.Context, n int) (result ChunkResult, err error) {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	selector := workflow.NewSelector(ctx)
	channel := workflow.NewChannel(ctx)
	wg := workflow.NewWaitGroup(ctx)

	workflow.Go(ctx, func(ctx workflow.Context) {
		for {
			var r ActivityNResult
			channel.Receive(ctx, &r)
			var res ActivityMResult
			err1 := workflow.ExecuteActivity(ctx, ActivityM, r.ID).Get(ctx, &res)
			if err1 != nil {
				err = err1
				return
			}
			wg.Done()
		}
	})

	for i := 0; i < n; i++ {
		future := workflow.ExecuteActivity(ctx, ActivityN, i+1)
		selector.AddFuture(future, func(f workflow.Future) {
			var r ActivityNResult
			err1 := f.Get(ctx, &r)
			if err1 != nil {
				err = err1
				return
			}

			wg.Add(1)
			workflow.Go(ctx, func(ctx workflow.Context) {
				channel.Send(ctx, r)
			})
		})
	}

	for i := 0; i < n; i++ {
		selector.Select(ctx)

		workflow.GetLogger(ctx).Info("the value of n is ", "n", n, "length", channel.Len())
	}
	wg.Wait(ctx)

	workflow.GetLogger(ctx).Info("Workflow completed.")
	return ChunkResult{1, 1}, nil
}

func ActivityN(ctx context.Context, ID int) (ActivityNResult, error) {
	time.Sleep(time.Second * time.Duration(rand.Int31n(5)))

	activity.GetLogger(ctx).Info("Activity N processed", "chunkID", ID)
	return ActivityNResult{
		ID: ID,
	}, nil
}

func ActivityM(ctx context.Context, ActivityNID int) (ActivityMResult, error) {
	time.Sleep(time.Second * 3)

	activity.GetLogger(ctx).Info("Activity M processed", "ActivityNID", ActivityNID)
	return ActivityMResult{}, nil
}
