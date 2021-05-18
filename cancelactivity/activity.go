package cancelactivity

import (
	"context"
	"time"

	"go.temporal.io/sdk/activity"
)

type Activities struct {
}

func (a *Activities) ActivityToBeCanceled(ctx context.Context) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("activity started, to cancel workflow, use 'go run cancelactivity/cancel/main.go " +
		"-w <WorkflowID>' or CLI: 'tctl wf cancel -w <WorkflowID>' to cancel")
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

func (a *Activities) CleanupActivity(ctx context.Context) error {
	logger := activity.GetLogger(ctx)
	logger.Info("cleanupActivity started")
	return nil
}

func (a *Activities) ActivityToBeSkipped(ctx context.Context) error {
	logger := activity.GetLogger(ctx)
	logger.Info("this activity will be skipped due to cancellation")
	return nil
}
