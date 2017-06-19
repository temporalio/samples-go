package main

import (
	"time"

	"go.uber.org/cadence"
	"go.uber.org/zap"
)

/**
 * This sample workflow executes multiple branches in parallel using cadence.Go() method.
 */

// This is registration process where you register all your workflows
// and activity function handlers.
func init() {
	cadence.RegisterWorkflow(SampleParallelWorkflow)
}

// SampleParallelWorkflow workflow decider
func SampleParallelWorkflow(ctx cadence.Context) error {
	waitChannel := cadence.NewChannel(ctx)

	ao := cadence.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
		HeartbeatTimeout:       time.Second * 20,
	}
	ctx = cadence.WithActivityOptions(ctx, ao)

	logger := cadence.GetLogger(ctx)
	cadence.Go(ctx, func(ctx cadence.Context) {
		err := cadence.ExecuteActivity(ctx, sampleActivity, "branch1.1").Get(ctx, nil)
		if err != nil {
			logger.Error("Activity failed", zap.Error(err))
		}
		err = cadence.ExecuteActivity(ctx, sampleActivity, "branch1.2").Get(ctx, nil)
		if err != nil {
			logger.Error("Activity failed", zap.Error(err))
		}
		waitChannel.Send(ctx, true)
	})

	cadence.Go(ctx, func(ctx cadence.Context) {
		err := cadence.ExecuteActivity(ctx, sampleActivity, "branch2").Get(ctx, nil)
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
