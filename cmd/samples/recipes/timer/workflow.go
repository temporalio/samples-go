package main

import (
	"context"
	"math/rand"
	"time"

	"go.uber.org/cadence"
	"go.uber.org/zap"
)

// ApplicationName is the task list for this sample
const ApplicationName = "timerGroup"

// This is registration process where you register all your workflows
// and activity function handlers.
func init() {
	cadence.RegisterWorkflow(SampleTimerWorkflow)
	cadence.RegisterActivity(orderProcessingActivity)
	cadence.RegisterActivity(sendEmailActivity)
}

// SampleTimerWorkflow workflow decider
func SampleTimerWorkflow(ctx cadence.Context, processingTimeThreshold time.Duration) error {
	ao := cadence.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
		HeartbeatTimeout:       time.Second * 20,
	}
	ctx = cadence.WithActivityOptions(ctx, ao)

	childCtx, cancelHandler := cadence.WithCancel(ctx)
	selector := cadence.NewSelector(ctx)

	// In this sample case, we want to demo a use case where the workflow starts a long running order processing operation
	// and in the case that the processing takes too long, we want to send out a notification email to user about the delay,
	// but we won't cancel the operation. If the operation finishes before the timer fires, then we want to cancel the timer.

	var processingDone bool
	f := cadence.ExecuteActivity(ctx, orderProcessingActivity)
	selector.AddFuture(f, func(f cadence.Future) {
		processingDone = true
		// cancel timerFuture
		cancelHandler()
	})

	// use timer future to send notification email if processing takes too long
	timerFuture := cadence.NewTimer(childCtx, processingTimeThreshold)
	selector.AddFuture(timerFuture, func(f cadence.Future) {
		if !processingDone {
			// processing is not done yet when timer fires, send notification email
			cadence.ExecuteActivity(ctx, sendEmailActivity).Get(ctx, nil)
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

	cadence.GetLogger(ctx).Info("Workflow completed.")
	return nil
}

func orderProcessingActivity(ctx context.Context) error {
	logger := cadence.GetActivityLogger(ctx)
	logger.Info("sampleActivity processing started.")
	timeNeededToProcess := time.Second * time.Duration(rand.Intn(10))
	time.Sleep(timeNeededToProcess)
	logger.Info("sampleActivity done.", zap.Duration("duration", timeNeededToProcess))
	return nil
}

func sendEmailActivity(ctx context.Context) error {
	cadence.GetActivityLogger(ctx).Info("sendEmailActivity sending notification email as the process takes long time.")
	return nil
}
