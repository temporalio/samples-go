package main

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/cadence"
)

/**
 * This sample workflow execute activities in parallel branches, pick the result of the branch that completes first,
 * and then cancels other activities that are not finished yet.
 */

// ApplicationName is the task list for this sample
const ApplicationName = "pickfirstGroup"

// This is registration process where you register all your workflows and activities
func init() {
	cadence.RegisterWorkflow(SamplePickFirstWorkflow)
	cadence.RegisterActivity(sampleActivity)
}

// SamplePickFirstWorkflow workflow decider
func SamplePickFirstWorkflow(ctx cadence.Context) error {
	selector := cadence.NewSelector(ctx)
	var firstResponse string

	// Use one cancel handler to cancel all of them. Cancelling on parent handler will close all the child ones
	// as well.
	childCtx, cancelHandler := cadence.WithCancel(ctx)
	ao := cadence.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
		HeartbeatTimeout:       time.Second * 20,
		WaitForCancellation:    true, // Wait for cancellation to complete.
	}
	childCtx = cadence.WithActivityOptions(ctx, ao)

	// Set WaitForCancellation to true to demonstrate the cancellation to the other activities. In real world case,
	// you might not care about them and could set WaitForCancellation to false (which is default value).

	// starts 2 activities in parallel
	f1 := cadence.ExecuteActivity(childCtx, sampleActivity, 0, time.Second*2)
	f2 := cadence.ExecuteActivity(childCtx, sampleActivity, 1, time.Second*10)
	pendingFutures := []cadence.Future{f1, f2}
	selector.AddFuture(f1, func(f cadence.Future) {
		f.Get(ctx, &firstResponse)
	}).AddFuture(f2, func(f cadence.Future) {
		f.Get(ctx, &firstResponse)
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
		f.Get(ctx, nil)
	}
	cadence.GetLogger(ctx).Info("Workflow completed.")
	return nil
}

func sampleActivity(ctx context.Context, currentBranchID int, totalDuration time.Duration) (string, error) {

	logger := cadence.GetActivityLogger(ctx)
	elapsedDuration := time.Nanosecond
	for elapsedDuration < totalDuration {
		time.Sleep(time.Second)
		elapsedDuration += time.Second

		// record heartbeat every second to check if we are been cancelled
		cadence.RecordActivityHeartbeat(ctx, "status-report-to-workflow")

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
