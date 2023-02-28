package periodic_sequence

import (
	"context"

	"github.com/temporalio/samples-go/polling"
)

type PollingActivities struct {
	TestService *polling.TestService
}

// DoPoll Activity.
func (a *PollingActivities) DoPoll(ctx context.Context) (string, error) {
	return a.TestService.GetServiceResult(ctx)
}
