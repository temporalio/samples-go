package greeting

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

// SampleWorkflow is a basic workflow definition
func SampleWorkflow(ctx workflow.Context, name string) (string, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("SampleWorkflow started", "name", name)

	// Configure activity options
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	// Execute the activity
	var result string
	err := workflow.ExecuteActivity(ctx, HelloActivity, name).Get(ctx, &result)
	if err != nil {
		logger.Error("Activity failed", "error", err)
		return "", err
	}

	logger.Info("SampleWorkflow completed", "result", result)
	return result, nil
}
