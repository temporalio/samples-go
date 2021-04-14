package retryactivity

import (
	"context"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

/**
 * This sample workflow executes unreliable activity with retry policy. If activity execution failed, server will
 * schedule retry based on retry policy configuration. The activity also heartbeat progress so it could resume from
 * reported progress in retry attempt.
 */

// RetryWorkflow workflow definition
func RetryWorkflow(ctx workflow.Context) error {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 2 * time.Minute,
		HeartbeatTimeout:    10 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    5,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	err := workflow.ExecuteActivity(ctx, BatchProcessingActivity, 0, 20, time.Second).Get(ctx, nil)
	if err != nil {
		workflow.GetLogger(ctx).Info("Workflow completed with error.", "Error", err)
		return err
	}
	workflow.GetLogger(ctx).Info("Workflow completed.")
	return nil
}

// BatchProcessingActivity process batchSize of jobs starting from firstTaskID. This activity will heartbeat to report
// progress, and it could fail sometimes. Use retry policy to retry when it failed, and resume from reported progress.
func BatchProcessingActivity(ctx context.Context, firstTaskID, batchSize int, processDelay time.Duration) error {
	logger := activity.GetLogger(ctx)

	i := firstTaskID
	if activity.HasHeartbeatDetails(ctx) {
		// we are retry from a failed attempt, and there is reported progress that we should resume from.
		var completedIdx int
		if err := activity.GetHeartbeatDetails(ctx, &completedIdx); err == nil {
			i = completedIdx + 1
			logger.Info("Resuming from failed attempt", "ReportedProgress", completedIdx)
		}
	}

	taskProcessedInThisAttempt := 0 // used to determine when to fail (simulate failure)
	for ; i < firstTaskID+batchSize; i++ {
		// process task i
		logger.Info("processing task", "TaskID", i)
		time.Sleep(processDelay) // simulate time spend on processing each task
		activity.RecordHeartbeat(ctx, i)
		taskProcessedInThisAttempt++

		// simulate failure after process 1/3 of the tasks
		if taskProcessedInThisAttempt >= batchSize/3 && i < firstTaskID+batchSize-1 {
			logger.Info("Activity failed, will retry...")
			// Activity could return *ApplicationError which is always retryable.
			// To return non-retryable error use temporal.NewNonRetryableApplicationError() constructor.
			return temporal.NewApplicationError("some retryable error", "SomeType")
		}
	}

	logger.Info("Activity succeed.")
	return nil
}
