// Package retryactivitynohb demonstrates an activity with a high failure rate and no heartbeating.
// Compare with the retryactivity sample: without heartbeating there is no progress saved between
// retries, so each attempt starts over from the beginning. This makes it useful for demonstrating
// activity pause/unpause on a retrying activity that has no internal state to resume from.
package retryactivitynohb

import (
	"context"
	"math/rand"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// RetryWorkflow executes BatchProcessingActivity with a retry policy and no attempt cap.
// The activity does not heartbeat, so retries always restart from the beginning.
// Use activity pause to stop the retry loop and unpause to resume it.
func RetryWorkflow(ctx workflow.Context) error {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
		// No HeartbeatTimeout — this activity does not heartbeat.
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    5 * time.Second,
			BackoffCoefficient: 1.0,
			MaximumInterval:    5 * time.Second,
			// No MaximumAttempts — retries indefinitely until paused or cancelled.
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	// Large batch ensures the activity never completes naturally; pause it to stop it.
	err := workflow.ExecuteActivity(ctx, BatchProcessingActivity, 0, 10, 2*time.Second).Get(ctx, nil)
	if err != nil {
		workflow.GetLogger(ctx).Info("Workflow completed with error.", "Error", err)
		return err
	}
	workflow.GetLogger(ctx).Info("Workflow completed.")
	return nil
}

// BatchProcessingActivity processes tasks one at a time, sleeping to simulate real work.
// Unlike the heartbeating variant, no progress is recorded between tasks, so each retry
// starts over from task 0 regardless of how far the previous attempt got.
// It always fails after 3 tasks, creating a high failure rate that keeps the retry loop going.
func BatchProcessingActivity(ctx context.Context, firstTaskID, batchSize int, processDelay time.Duration) error {
	logger := activity.GetLogger(ctx)

	for i := firstTaskID; i < firstTaskID+batchSize; i++ {
		// // Inject a 95% failure rate before doing any work on this task.
		if rand.Intn(100) < 95 {
			logger.Info("Simulating transient failure", "TaskID", i)
			return temporal.NewApplicationError("transient error", "SomeType")
		}

		logger.Info("Processing task", "TaskID", i)
		time.Sleep(processDelay)
	}

	logger.Info("Activity succeeded.")
	return nil
}
