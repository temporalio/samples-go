package frequent

import (
	"context"
	"time"

	"github.com/temporalio/samples-go/polling"
	"go.temporal.io/sdk/activity"
)

type PollingActivities struct {
	TestService  *polling.TestService
	PollInterval time.Duration
}

// DoPoll Activity.
// In this activity polling is implemented within the activity itself and not the workflow,
// using the heartbeat mechanism to keep the activity alive
// This activity intentionally hides underlying error and always retry until the context is closed
// A more sophisticated implementation would distinguish between intermittent failures and catastrophic
// failures from the underlying service
func (a *PollingActivities) DoPoll(ctx context.Context) (string, error) {
	for {
		res, err := a.TestService.GetServiceResult(ctx)
		if err == nil {
			return res, err
		}
		activity.RecordHeartbeat(ctx)
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(a.PollInterval):
		}
	}
}
