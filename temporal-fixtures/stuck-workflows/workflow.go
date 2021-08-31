package stuckworkflows

import (
	"context"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

/**
 * This sample workflow executes unreliable activity with retry policy. If activity execution failed, server will
 * schedule retry based on retry policy configuration. The activity also heartbeatы progress so it could resume from
 * reported progress in the retry attempts.
 */

// StuckWorkflow workflow definition
func StuckWorkflow(ctx workflow.Context) error {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 24 * time.Hour,
		HeartbeatTimeout:    10 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    0,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	err := workflow.ExecuteActivity(ctx, StuckWorkflowActivity).Get(ctx, nil)
	if err != nil {
		workflow.GetLogger(ctx).Info("Workflow completed with error.", "Error", err)
		return err
	}
	workflow.GetLogger(ctx).Info("Workflow completed.")
	return nil
}

// StuckWorkflowActivity process batchSize of jobs starting from firstTaskID. This activity will heartbeat to report
// progress, and it could fail sometimes. Use retry policy to retry when it failed, and resume from reported progress.
func StuckWorkflowActivity(ctx context.Context) error {
	logger := activity.GetLogger(ctx)

	logger.Info("Activity failed, will retry...")
	return temporal.NewApplicationError("some retryable error", "SomeType")
}
