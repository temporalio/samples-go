// @@@SNIPSTART samples-go-nexus-standalone-activity-handler
package handler

import (
	"context"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporalnexus"

	"github.com/temporalio/samples-go/nexus-standalone-activity/service"
)

// CreateGreetingActivity builds the greeting. It runs as a Standalone Activity, so it does not
// need a backing Workflow.
func CreateGreetingActivity(ctx context.Context, input service.GreetingInput) (service.GreetingOutput, error) {
	activity.GetLogger(ctx).Info("CreateGreeting", "name", input.Name)
	return service.GreetingOutput{Message: "Hello, " + input.Name + "!"}, nil
}

// GreetOperation is the Nexus Operation backed by CreateGreetingActivity. Start dispatches the
// Activity and returns its handle, so the caller's Nexus Operation completes when the Activity does.
var GreetOperation = temporalnexus.MustNewTemporalOperation(
	temporalnexus.TemporalOperationOptions[service.GreetingInput, service.GreetingOutput]{
		Name: service.GreetOperationName,
		Start: func(
			ctx context.Context,
			nc temporalnexus.NexusClient,
			input service.GreetingInput,
			options temporalnexus.StartTemporalOperationOptions,
		) (temporalnexus.TemporalOperationResult[service.GreetingOutput], error) {
			return temporalnexus.StartActivity(ctx, nc, client.StartActivityOptions{
				ID:                  service.GreetingActivityID(input),
				StartToCloseTimeout: 10 * time.Second,
			}, CreateGreetingActivity, input)
		},
	},
)

// @@@SNIPEND
