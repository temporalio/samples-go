package retryactivity

import (
	"context"
	"math/rand"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// RetryWorkflow executes BatchProcessingActivity with a retry policy and no attempt cap.
// The activity heartbeats progress after each task so retries resume from where they left off,
// rather than starting over. This makes it suitable for demonstrating activity pause/unpause:
// pausing mid-execution shows the last heartbeated task index in the UI, and unpausing
// resumes from that point.
func RetryWorkflow(ctx workflow.Context) error {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
		HeartbeatTimeout:    10 * time.Second,
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
// After each task it heartbeats the task index as progress. On retry the activity resumes
// from the last heartbeated index rather than starting over.
// It always fails after 3 tasks, creating a high failure rate that keeps the retry loop going.
func BatchProcessingActivity(ctx context.Context, firstTaskID, batchSize int, processDelay time.Duration) error {
	logger := activity.GetLogger(ctx)

	i := firstTaskID
	if activity.HasHeartbeatDetails(ctx) {
		// Resume from reported progress on retry.
		var completedIdx int
		if err := activity.GetHeartbeatDetails(ctx, &completedIdx); err == nil {
			i = completedIdx + 1
			logger.Info("Resuming from previous attempt", "ResumedAt", i)
		}
	}

	taskProcessedInThisAttempt := 0
	for ; i < firstTaskID+batchSize; i++ {
		// Inject a 95% failure rate before doing any work on this task.
		if rand.Intn(100) < 95 {
			logger.Info("Simulating transient failure", "TaskID", i)
			time.Sleep(5 * time.Second)
			return temporal.NewApplicationError("transient error", "SomeType")
		}

		logger.Info("Processing task", "TaskID", i)
		time.Sleep(processDelay)
		activity.RecordHeartbeat(ctx, i)
		taskProcessedInThisAttempt++
	}

	logger.Info("Activity succeeded.")
	return nil
}
