package infrequent

import (
	"github.com/temporalio/samples-go/polling"
)

// @@@SNIPSTART samples-go-polling-infrequent-activities
type PollingActivities struct {
	TestService *polling.TestService
}

// DoPoll Activity.
func (a *PollingActivities) DoPoll() (string, error) {
	return a.TestService.GetServiceResult()
}

// @@@SNIPEND
