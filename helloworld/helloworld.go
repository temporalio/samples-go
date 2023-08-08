package helloworld

import (
	"context"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/workflow"
	// TODO(cretz): Remove when tagged
	_ "go.temporal.io/sdk/contrib/tools/workflowcheck/determinism"
)

// Workflow is a Hello World workflow definition.
func Workflow(ctx workflow.Context, name string) (string, error) {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("HelloWorld workflow started", "name", name)

	var result string
	f1 := workflow.ExecuteActivity(ctx, Activity, name)
	selector := workflow.NewSelector(ctx)
	selector.AddFuture(f1, func(f workflow.Future) {
		err := f.Get(ctx, &result)
		if err != nil {
			logger.Error("Activity failed.", "Error", err)
		}
	})
	println(workflow.Now(ctx).String())
	selector.Select(ctx)
	// _ = workflow.Sleep(ctx, time.Millisecond)
	println(workflow.Now(ctx).String())

	logger.Info("HelloWorld workflow completed.", "result", result)

	return result, nil
}

func Activity(ctx context.Context, name string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Activity", "name", name)
	return "Hello " + name + "!", nil
}
