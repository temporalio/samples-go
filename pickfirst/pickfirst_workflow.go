package pickfirst

import (
	"context"
	"fmt"
	"time"

	"go.temporal.io/temporal/activity"
	"go.temporal.io/temporal/workflow"
)

/**
 * This sample workflow execute activities in parallel branches, pick the result of the branch that completes first,
 * and then cancels other activities that are not finished yet.
 */

// SamplePickFirstWorkflow workflow decider
func SamplePickFirstWorkflow(ctx workflow.Context) error {
	selector := workflow.NewSelector(ctx)
	var firstResponse string

	// Use one cancel handler to cancel all of them. Cancelling on parent handler will close all the child ones
	// as well.
	childCtx, cancelHandler := workflow.WithCancel(ctx)
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
		HeartbeatTimeout:       time.Second * 20,
		WaitForCancellation:    true, // Wait for cancellation to complete.
	}
	childCtx = workflow.WithActivityOptions(childCtx, ao)

	// Set WaitForCancellation to true to demonstrate the cancellation to the other activities. In real world case,
	// you might not care about them and could set WaitForCancellation to false (which is default value).

	// starts 2 activities in parallel
	f1 := workflow.ExecuteActivity(childCtx, SampleActivity, 0, time.Second*2)
	f2 := workflow.ExecuteActivity(childCtx, SampleActivity, 1, time.Second*10)
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
