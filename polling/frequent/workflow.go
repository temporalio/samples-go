package frequent

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

func FrequentPolling(ctx workflow.Context) (string, error) {
	logger := workflow.GetLogger(ctx)

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
		HeartbeatTimeout:    2 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	// @@@SNIPSTART samples-go-polling-frequent-activity
	var a *PollingActivities // use a nil struct pointer to call activities that are part of a structure

	var pollResult string
	err := workflow.ExecuteActivity(ctx, a.DoPoll).Get(ctx, &pollResult)
	if err != nil {
		logger.Error("Get greeting failed.", "Error", err)
		return "", err
	}
	// @@@SNIPEND

	return pollResult, err
}
