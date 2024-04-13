package asyncactivity

import (
	"context"
	"fmt"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/workflow"
)

// AsyncActivityWorkflow is a workflow definition starting two activities
// asynchronously.
func AsyncActivityWorkflow(ctx workflow.Context, name string) (string, error) {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	// Start activities asynchronously.
	var helloResult, byeResult string
	helloFuture := workflow.ExecuteActivity(ctx, HelloActivity, name)
	byeFuture := workflow.ExecuteActivity(ctx, ByeActivity, name)

	// This can be done alternatively by creating a workflow selector. See
	// "pickfirst" example.
	err := helloFuture.Get(ctx, &helloResult)
	if err != nil {
		return "", fmt.Errorf("hello activity error: %s", err.Error())
	}
	err = byeFuture.Get(ctx, &byeResult)
	if err != nil {
		return "", fmt.Errorf("bye activity error: %s", err.Error())
	}

	return helloResult + "\n" + byeResult, nil
}

// Each of these activities will sleep for 5 seconds, but see in the temporal
// dashboard that they were created immediately one after the other.

func HelloActivity(ctx context.Context, name string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Hello activity", "name", name)
	time.Sleep(5 * time.Second)
	return "Hello " + name + "!", nil
}

func ByeActivity(ctx context.Context, name string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Bye activity", "name", name)
	time.Sleep(5 * time.Second)
	return "Bye " + name + "!", nil
}
