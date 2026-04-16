package app

import (
	"context"
	"fmt"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/workflow"
)

// GreetingWorkflow orchestrates the greeting process.
func GreetingWorkflow(ctx workflow.Context, name string) (string, error) {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var result string
	err := workflow.ExecuteActivity(ctx, ComposeGreeting, name).Get(ctx, &result)
	if err != nil {
		return "", err
	}

	return result, nil
}

// ComposeGreeting is an activity that builds the greeting string.
func ComposeGreeting(ctx context.Context, name string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("ComposeGreeting activity called", "name", name)
	return fmt.Sprintf("Hello, %s!", name), nil
}
