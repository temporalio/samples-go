package frequent

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

const (
	TaskQueueName = "pollingFrequentlySampleQueue"
)

func FrequentPolling(ctx workflow.Context) (string, error) {
	logger := workflow.GetLogger(ctx)

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 24 * time.Hour,
		HeartbeatTimeout:    30 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var a *PollingActivities // use a nil struct pointer to call activities that are part of a structure

	var pollResult string
	err := workflow.ExecuteActivity(ctx, a.DoPoll).Get(ctx, &pollResult)
	if err != nil {
		logger.Error("Frequent polling activity failed.", "Error", err)
		return "", err
	}

	return pollResult, err
}
