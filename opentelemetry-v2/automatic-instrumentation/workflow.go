package automaticinstrumentation

import (
	"context"
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"
)

const TaskQueueName = "otel-v2-automatic"

func Workflow(ctx workflow.Context, name string) (string, error) {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	})

	var result string
	if err := workflow.ExecuteActivity(ctx, Activity, name).Get(ctx, &result); err != nil {
		return "", err
	}

	return result, nil
}

func Activity(ctx context.Context, name string) (string, error) {
	return fmt.Sprintf("Hello, %s!", name), nil
}
