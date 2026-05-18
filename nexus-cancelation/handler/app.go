package handler

import (
	"context"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/nexus-rpc/sdk-go/nexus"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/temporalnexus"
	"go.temporal.io/sdk/workflow"

	"github.com/temporalio/samples-go/nexus/service"
)

// Use the NewWorkflowRunOperation constructor, which is the easiest way to expose a workflow as an operation.
// See alternatives at https://pkg.go.dev/go.temporal.io/sdk/temporalnexus.
var HelloOperation = temporalnexus.NewWorkflowRunOperation(service.HelloOperationName, HelloHandlerWorkflow, func(ctx context.Context, input service.HelloInput, options nexus.StartOperationOptions) (client.StartWorkflowOptions, error) {
	return client.StartWorkflowOptions{
		// Workflow IDs should typically be business meaningful IDs and are used to dedupe workflow starts.
		// For this example, use a business ID derived from the greeting input so repeated operations
		// for the same name and language resolve to the same workflow.
		ID: service.HelloWorkflowID(input),
		// Task queue defaults to the task queue this operation is handled on.
	}, nil
})

func HelloHandlerWorkflow(ctx workflow.Context, input service.HelloInput) (service.HelloOutput, error) {
	// Sleep for a random duration to simulate some work
	var duration time.Duration
	err := workflow.SideEffect(ctx, func(ctx workflow.Context) any {
		return time.Duration(rand.IntN(5)) * time.Second
	}).Get(&duration)
	if err != nil {
		return service.HelloOutput{}, err
	}
	if err := workflow.Sleep(ctx, duration); err != nil {
		// Simulate some work after cancellation is requested
		sleepErr := err
		if temporal.IsCanceledError(err) {
			ctx, _ = workflow.NewDisconnectedContext(ctx)
			var duration time.Duration
			err := workflow.SideEffect(ctx, func(ctx workflow.Context) any {
				return time.Duration(rand.IntN(5)) * time.Second
			}).Get(&duration)
			if err != nil {
				return service.HelloOutput{}, err
			}
			if err := workflow.Sleep(ctx, duration); err != nil {
				return service.HelloOutput{}, err
			}
		}
		return service.HelloOutput{}, sleepErr
	}

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
