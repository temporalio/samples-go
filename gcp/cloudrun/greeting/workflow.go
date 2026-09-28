package greeting

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

// SampleWorkflow executes a single greeting activity.
func SampleWorkflow(ctx workflow.Context, name string) (string, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("SampleWorkflow started", "name", name)

	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	})

	var result string
	if err := workflow.ExecuteActivity(ctx, HelloActivity, name).Get(ctx, &result); err != nil {
		logger.Error("Activity failed", "error", err)
		return "", err
	}

	logger.Info("SampleWorkflow completed", "result", result)
	return result, nil
}
