package frequent

import (
	"context"
	"errors"
	"github.com/temporalio/samples-go/polling"
	"go.temporal.io/sdk/activity"
	"time"
)

type PollingActivities struct {
	TestService  *polling.TestService
	PollInterval time.Duration
}

// DoPoll Activity.
// In this activity polling is implemented within the activity itself and not the workflow,
// using the heartbeat mechanism to keep the activity alive
func (a *PollingActivities) DoPoll(ctx context.Context) (string, error) {
	for {
		res, err := a.TestService.GetServiceResult(ctx)
		if err == nil {
			return res, err
		}
		activity.RecordHeartbeat(ctx)
		select {
		case <-ctx.Done():
			return "", errors.New("channel closed")
		case <-time.After(a.PollInterval):
			return a.TestService.GetServiceResult(ctx)
		}
	}
}
