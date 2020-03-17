package main

import (
	"time"

	"go.temporal.io/temporal/workflow"
	"go.uber.org/zap"
)

/**
 * This sample workflow executes multiple branches in parallel using workflow.Go() method.
 */

// SampleParallelWorkflow workflow decider
func SampleParallelWorkflow(ctx workflow.Context) error {
	waitChannel := workflow.NewChannel(ctx)

	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
		HeartbeatTimeout:       time.Second * 20,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	workflow.Go(ctx, func(ctx workflow.Context) {
		err := workflow.ExecuteActivity(ctx, sampleActivity, "branch1.1").Get(ctx, nil)
		if err != nil {
			logger.Error("Activity failed", zap.Error(err))
		}
		err = workflow.ExecuteActivity(ctx, sampleActivity, "branch1.2").Get(ctx, nil)
		if err != nil {
			logger.Error("Activity failed", zap.Error(err))
		}
		waitChannel.Send(ctx, true)
	})

	workflow.Go(ctx, func(ctx workflow.Context) {
		err := workflow.ExecuteActivity(ctx, sampleActivity, "branch2").Get(ctx, nil)
		if err != nil {
			logger.Error("Activity failed", zap.Error(err))
		}
		waitChannel.Send(ctx, true)
	})

	// wait for both of the coroutinue to complete.
	waitChannel.Receive(ctx, nil)
	waitChannel.Receive(ctx, nil)
	logger.Info("Workflow completed.")
	return nil
}
