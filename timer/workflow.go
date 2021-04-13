package timer

import (
	"context"
	"math/rand"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/workflow"
)

// SampleTimerWorkflow workflow definition
func SampleTimerWorkflow(ctx workflow.Context, processingTimeThreshold time.Duration) error {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	childCtx, cancelHandler := workflow.WithCancel(ctx)
	selector := workflow.NewSelector(ctx)

	// In this sample case, we want to demo a use case where the workflow starts a long running order processing operation
	// and in the case that the processing takes too long, we want to send out a notification email to user about the delay,
	// but we won't cancel the operation. If the operation finishes before the timer fires, then we want to cancel the timer.

	var processingDone bool
	f := workflow.ExecuteActivity(ctx, OrderProcessingActivity)
	selector.AddFuture(f, func(f workflow.Future) {
		processingDone = true
		// cancel timerFuture
		cancelHandler()
	})

	// use timer future to send notification email if processing takes too long
	timerFuture := workflow.NewTimer(childCtx, processingTimeThreshold)
	selector.AddFuture(timerFuture, func(f workflow.Future) {
		if !processingDone {
			// processing is not done yet when timer fires, send notification email
			_ = workflow.ExecuteActivity(ctx, SendEmailActivity).Get(ctx, nil)
		}
	})

	// wait the timer or the order processing to finish
	selector.Select(ctx)

	// now either the order processing is finished, or timer is fired.
	if !processingDone {
		// processing not done yet, so the handler for timer will send out notification email.
		// we still want the order processing to finish, so wait on it.
		selector.Select(ctx)
	}

	workflow.GetLogger(ctx).Info("Workflow completed.")
	return nil
}

func OrderProcessingActivity(ctx context.Context) error {
	logger := activity.GetLogger(ctx)
	logger.Info("OrderProcessingActivity processing started.")
	timeNeededToProcess := time.Second * time.Duration(rand.Intn(10))
	time.Sleep(timeNeededToProcess)
	logger.Info("OrderProcessingActivity done.", "duration", timeNeededToProcess)
	return nil
}

func SendEmailActivity(ctx context.Context) error {
	activity.GetLogger(ctx).Info("SendEmailActivity sending notification email as the process takes long time.")
	return nil
}
