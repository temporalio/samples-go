package periodic_sequence

import (
	"time"

	"go.temporal.io/sdk/temporal"

	"go.temporal.io/sdk/workflow"
)

type ChildWorkflowParams struct {
	PollingInterval time.Duration
}

func PollingChildWorkflow(ctx workflow.Context, params ChildWorkflowParams) (string, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting child workflow with params", params)
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 4 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 1,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var a *PollingActivities

	for i := 0; i < 10; i++ {
		// Here we would invoke a sequence of activities
		// For sample we just use a single one repeated several times
		var pollResult string
		err := workflow.ExecuteActivity(ctx, a.DoPoll).Get(ctx, &pollResult)
		if err == nil {
			return pollResult, nil
		}
		logger.Error("Error in activity, sleeping and retrying", err)
		err = workflow.Sleep(ctx, params.PollingInterval)
		if err != nil {
			return "", err
		}
	}
	// Request that the new child workflow run is invoked
	err := workflow.NewContinueAsNewError(ctx, PollingChildWorkflow, params)
	return "", err
}
