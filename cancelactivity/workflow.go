package cancelactivity

import (
	"context"
	"fmt"
	"time"

	"go.temporal.io/temporal/activity"
	"go.temporal.io/temporal/workflow"
	"go.uber.org/zap"
)

/**
 * This is the cancel activity workflow sample.
 */

// Workflow workflow
func Workflow(ctx workflow.Context) error {
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute * 30,
		HeartbeatTimeout:       time.Second * 5,
		WaitForCancellation:    true,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	logger := workflow.GetLogger(ctx)
	logger.Info("cancel workflow started")

	defer func() {
		// When workflow is canceled, it has to get a new disconnected context to execute any activities
		newCtx, _ := workflow.NewDisconnectedContext(ctx)
		err := workflow.ExecuteActivity(newCtx, CleanupActivity).Get(ctx, nil)
		if err != nil {
			logger.Error("Cleanup activity failed", zap.Error(err))
		}
	}()

	var result string
	err := workflow.ExecuteActivity(ctx, ActivityToBeCanceled).Get(ctx, &result)
	logger.Info(fmt.Sprintf("activityToBeCanceled returns %v, %v", result, err))

	err = workflow.ExecuteActivity(ctx, ActivityToBeSkipped).Get(ctx, nil)
	logger.Error("Error from activityToBeSkipped", zap.Error(err))

	logger.Info("Workflow completed.")

	return nil
}

func ActivityToBeCanceled(ctx context.Context) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("activity started, to cancel workflow, use 'go run cancelactivity/cancel/main.go -w <WorkflowID>' or CLI: 'tctl wf cancel -w <WorkflowID>' to cancel")
	for {
		select {
		case <-time.After(1 * time.Second):
			logger.Info("heartbeating...")
			activity.RecordHeartbeat(ctx, "")
		case <-ctx.Done():
			logger.Info("context is cancelled")
			return "I am canceled by Done", nil
		}
	}
}

func CleanupActivity(ctx context.Context) error {
	logger := activity.GetLogger(ctx)
	logger.Info("cleanupActivity started")
	return nil
}

func ActivityToBeSkipped(ctx context.Context) error {
	logger := activity.GetLogger(ctx)
	logger.Info("this activity will be skipped due to cancellation")
	return nil
}
