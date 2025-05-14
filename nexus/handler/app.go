// @@@SNIPSTART samples-go-nexus-handler
package handler

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/nexus-rpc/sdk-go/nexus"

	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/temporalnexus"
	"go.temporal.io/sdk/workflow"

	"github.com/temporalio/samples-go/nexus/service"
)

// NewSyncOperation is a meant for exposing simple RPC handlers.
var EchoOperation = nexus.NewSyncOperation(service.EchoOperationName, func(ctx context.Context, input service.EchoInput, options nexus.StartOperationOptions) (service.EchoOutput, error) {
	// Use temporalnexus.GetClient to get the client that the worker was initialized with to perform client calls
	// such as signaling, querying, and listing workflows. Implementations are free to make arbitrary calls to other
	// services or databases, or perform simple computations such as this one.
	return service.EchoOutput{
		Message: "Hello " + input.Message,
	}, nil
})

// Use the NewWorkflowRunOperation constructor, which is the easiest way to expose a workflow as an operation.
// See alternatives at https://pkg.go.dev/go.temporal.io/sdk/temporalnexus.
var HelloOperation = temporalnexus.NewWorkflowRunOperation(service.HelloOperationName, HelloHandlerWorkflow, func(ctx context.Context, input service.HelloInput, options nexus.StartOperationOptions) (client.StartWorkflowOptions, error) {
	return client.StartWorkflowOptions{
		ID: "hello-" + input.Name,
		// Attach to existing workflow if it's already running.
		WorkflowIDConflictPolicy: enumspb.WORKFLOW_ID_CONFLICT_POLICY_USE_EXISTING,
		// Task queue defaults to the task queue this operation is handled on.
		WorkflowRunTimeout: 15000 * time.Millisecond,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    1 * time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    100 * time.Second,
			MaximumAttempts:    0,
		},
	}, nil
})

var CancelOperation = temporalnexus.NewWorkflowRunOperation(
	service.CancelOperationName,
	CancelHandlerWorkflow,
	func(ctx context.Context, _ nexus.NoValue, options nexus.StartOperationOptions) (client.StartWorkflowOptions, error) {
		return client.StartWorkflowOptions{
			ID:                       options.RequestID,
			WorkflowExecutionTimeout: 5 * time.Minute,
		}, nil
	},
)

func HelloHandlerWorkflow(ctx workflow.Context, input service.HelloInput) (service.HelloOutput, error) {
	var sleepDuration time.Duration
	encodedValue := workflow.SideEffect(
		ctx,
		func(ctx workflow.Context) interface{} { return time.Duration(rand.Int63n(10)+1) * time.Second },
	)
	encodedValue.Get(&sleepDuration)

	fmt.Printf("Sleeping for %s\n", sleepDuration)
	if err := workflow.Sleep(ctx, sleepDuration); err != nil {
		return service.HelloOutput{}, nil
	}

	if sleepDuration <= 5*time.Second {
		return service.HelloOutput{}, workflow.NewContinueAsNewError(ctx, HelloHandlerWorkflow, input)
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

func CancelHandlerWorkflow(ctx workflow.Context, _ nexus.NoValue) (nexus.NoValue, error) {
	return nil, workflow.Await(ctx, func() bool { return false })
}

// @@@SNIPEND
