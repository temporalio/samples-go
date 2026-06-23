package handler

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporalnexus"

	"github.com/temporalio/samples-go/nexus-activity-operation/service"
)

// HelloOperation is an asynchronous Nexus operation backed by a standalone activity execution.
// The Start callback uses temporalnexus.StartActivity to schedule HelloHandlerActivity and
// returns an async result whose operation token resolves when the activity completes.
//
// Activity-backed operations skip the cost of a workflow execution when the underlying work is
// a single side-effecting call (an API request, a database write, a compute step). They retain
// Temporal's retry, timeout, and cancellation semantics via standard activity options.
var HelloOperation = temporalnexus.MustNewTemporalOperation(temporalnexus.TemporalOperationOptions[service.HelloInput, service.HelloOutput]{
	Name: service.HelloOperationName,
	Start: func(ctx context.Context, nc temporalnexus.NexusClient, input service.HelloInput, _ temporalnexus.StartTemporalOperationOptions) (temporalnexus.TemporalOperationResult[service.HelloOutput], error) {
		return temporalnexus.StartActivity(ctx, nc, client.StartActivityOptions{
			ID:                  helloActivityID(input),
			StartToCloseTimeout: 30 * time.Second,
			// TaskQueue defaults to the task queue this operation is handled on.
		}, HelloHandlerActivity, input)
	},
})

// helloActivityID returns a deterministic activity ID derived from the input. The Nexus start
// request may be retried, and a non-deterministic ID would embed a fresh value into the
// operation token on each retry.
func helloActivityID(input service.HelloInput) string {
	name := strings.ToLower(strings.ReplaceAll(strings.TrimSpace(input.Name), " ", "-"))
	return fmt.Sprintf("hello-%s-%s", input.Language, name)
}

func HelloHandlerActivity(_ context.Context, input service.HelloInput) (service.HelloOutput, error) {
	switch input.Language {
	case service.EN:
		return service.HelloOutput{Message: "Hello " + input.Name + " 👋"}, nil
	case service.FR:
		return service.HelloOutput{Message: "Bonjour " + input.Name + " 👋"}, nil
	case service.DE:
		return service.HelloOutput{Message: "Hallo " + input.Name + " 👋"}, nil
	case service.ES:
		return service.HelloOutput{Message: "¡Hola! " + input.Name + " 👋"}, nil
	case service.TR:
		return service.HelloOutput{Message: "Merhaba " + input.Name + " 👋"}, nil
	}
	return service.HelloOutput{}, fmt.Errorf("unsupported language %q", input.Language)
}
