package pickfirst

import (
	"context"
	"fmt"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/workflow"
)

/**
 * This sample workflow execute activities in parallel branches, pick the result of the branch that completes first,
 * and then cancels other activities that are not finished yet.
 */

// SamplePickFirstWorkflow workflow definition
func SamplePickFirstWorkflow(ctx workflow.Context) error {
	selector := workflow.NewSelector(ctx)
	var firstResponse string

	// Use one cancel handler to cancel all of them. Cancelling on parent handler will close all the child ones
	// as well.
	childCtx, cancelHandler := workflow.WithCancel(ctx)
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 2 * time.Minute,
		HeartbeatTimeout:    10 * time.Second,
		WaitForCancellation: true, // Wait for cancellation to complete.
	}
	childCtx = workflow.WithActivityOptions(childCtx, ao)

	// Set WaitForCancellation to true to demonstrate the cancellation to the other activities. In real world case,
	// you might not care about them and could set WaitForCancellation to false (which is default value).

	// starts 2 activities in parallel
	// Duration of f1 is set to 10 seconds in order to observe the cancellation before timeout, because
	// Cancel is not delivered to activity until a heartbeat has been actually sent.
	// Due to the internal batching, the first heartbeat will not be sent until after 80% of the HeartbeatTimeout (8 seconds in this case).
	f1 := workflow.ExecuteActivity(childCtx, SampleActivity, 0, 10*time.Second)
	f2 := workflow.ExecuteActivity(childCtx, SampleActivity, 1, 1*time.Second)
	pendingFutures := []workflow.Future{f1, f2}
	selector.AddFuture(f1, func(f workflow.Future) {
		_ = f.Get(ctx, &firstResponse)
	}).AddFuture(f2, func(f workflow.Future) {
		_ = f.Get(ctx, &firstResponse)
	})

	// wait for any of the future to complete
	selector.Select(ctx)

	// now at least one future is complete, so cancel all other pending futures.
	cancelHandler()

	// - If you want to wait for pending activities to finish after issuing cancellation
	// then wait for the future to complete.
	// - if you don't want to wait for completion of pending activities cancellation then you can choose to
	// set WaitForCancellation to false through WithWaitForCancellation(false)
	for _, f := range pendingFutures {
		_ = f.Get(ctx, nil)
	}
	workflow.GetLogger(ctx).Info("Workflow completed.")
	return nil
}

func SampleActivity(ctx context.Context, currentBranchID int, totalDuration time.Duration) (string, error) {

	logger := activity.GetLogger(ctx)
	elapsedDuration := time.Nanosecond
	for elapsedDuration < totalDuration {
		time.Sleep(time.Second)
		elapsedDuration += time.Second

		// record heartbeat every second to check if we are been cancelled
		activity.RecordHeartbeat(ctx, "status-report-to-workflow")

		select {
		case <-ctx.Done():
			// We have been cancelled.
			msg := fmt.Sprintf("Branch %d is cancelled.", currentBranchID)
			logger.Info(msg)
			return msg, ctx.Err()
		default:
			// We are not cancelled yet.
		}

		// Do some custom work
		// ...
	}

	msg := fmt.Sprintf("Branch %d done in %s.", currentBranchID, totalDuration)
	return msg, nil
}
