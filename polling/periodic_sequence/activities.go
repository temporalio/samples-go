package periodic_sequence

import (
	"github.com/temporalio/samples-go/polling"
)

// @@@SNIPSTART samples-go-polling-periodic-sequence
type PollingActivities struct {
	TestService *polling.TestService
}

// DoPoll Activity.
func (a *PollingActivities) DoPoll() (string, error) {
	return a.TestService.GetServiceResult()
}

// @@@SNIPEND
