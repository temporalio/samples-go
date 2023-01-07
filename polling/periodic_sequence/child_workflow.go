package periodic_sequence

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

type ChildWorkflowParams struct {
	SingleWorkflowAttempts int
	PollingInterval        time.Duration
}

func ChildWorkflow(ctx workflow.Context, params ChildWorkflowParams) (string, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting child workflow with params", params)
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var a *PollingActivities

	for i := 0; i < params.SingleWorkflowAttempts; i++ {
		// Here we would invoke a sequence of activities
		// For sample we just use a single one repeated several times
		var pollResult string
		err := workflow.ExecuteActivity(ctx, a.DoPoll).Get(ctx, &pollResult)
		if err == nil {
			return pollResult, nil
		} else {
			logger.Error("Error in activity, sleeping and retrying", err)
			workflow.Sleep(ctx, params.PollingInterval)
		}
	}
	// Request that the new child workflow run is invoked
	err := workflow.NewContinueAsNewError(ctx, ChildWorkflow, params)
	return "", err
}
