package clientupdate

import (
	"fmt"

	temporalotel "go.temporal.io/sdk/contrib/opentelemetry-v2"
	"go.temporal.io/sdk/workflow"
)

const (
	instrumentationName = "github.com/temporalio/samples-go/opentelemetry-v2/custom-instrumentation/client-update-propagation"
	TaskQueueName       = "otel-v2-custom-client-update-propagation"
	UpdateName          = "greet"
)

// @@@SNIPSTART samples-go-opentelemetry-v2-update-handler-span
func Workflow(ctx workflow.Context) (string, error) {
	var result string
	updateCompleted := false

	err := workflow.SetUpdateHandlerWithOptions(
		ctx,
		UpdateName,
		func(ctx workflow.Context, name string) (string, error) {
			_, span := temporalotel.Tracer(instrumentationName).Start(ctx, "handle-update")
			defer span.End()

			result = fmt.Sprintf("Hello, %s!", name)
			updateCompleted = true
			return result, nil
		},
		workflow.UpdateHandlerOptions{
			Validator: func(ctx workflow.Context, name string) error {
				_, span := temporalotel.Tracer(instrumentationName).Start(ctx, "validate-update")
				defer span.End()

				if name == "" {
					return fmt.Errorf("name cannot be empty")
				}
				return nil
			},
		},
	)
	if err != nil {
		return "", fmt.Errorf("unable to register Update handler: %w", err)
	}

	if err := workflow.Await(ctx, func() bool { return updateCompleted && workflow.AllHandlersFinished(ctx) }); err != nil {
		return "", err
	}
	return result, nil
}

// @@@SNIPEND
