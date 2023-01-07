package infrequent

import (
	"go.temporal.io/sdk/temporal"
	"time"

	"go.temporal.io/sdk/workflow"
)

// InfrequentPolling Workflow that shows how infrequent polling via activity can be
// implemented via activity retries. For this sample we  want to poll the test service
// every 60 seconds. Here we:
//
//  1. Set RetryPolicy backoff coefficient of 1
//  2. Set initial interval to the poll frequency (since coefficient is 1, same interval will
//     be used for all retries)
//
// With this in case our test service is "down" we can fail our activity, and it will be
// retried based on our 60 seconds retry interval until poll is successful, and we can return a
// result from the activity.
func InfrequentPolling(ctx workflow.Context) (string, error) {
	logger := workflow.GetLogger(ctx)
	// @@@SNIPSTART samples-go-polling-infrequent-activities
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 2 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			BackoffCoefficient: 1,
			InitialInterval:    60 * time.Second,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	// @@@SNIPEND

	var a *PollingActivities // use a nil struct pointer to call activities that are part of a structure

	var pollResult string
	err := workflow.ExecuteActivity(ctx, a.DoPoll).Get(ctx, &pollResult)
	if err != nil {
		logger.Error("Polling failed.", "Error", err)
		return "", err
	}

	return pollResult, nil
}
