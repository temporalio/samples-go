// @@@SNIPSTART samples-go-nexus-handler
package handler

import (
	"context"
	"fmt"
	"time"

	"github.com/nexus-rpc/sdk-go/nexus"

	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporalnexus"
	"go.temporal.io/sdk/workflow"

	"github.com/temporalio/samples-go/nexus/service"
)

// NewSyncOperation is a meant for exposing simple RPC handlers.
var EchoOperation = temporalnexus.NewSyncOperation(service.EchoOperationName, func(ctx context.Context, c client.Client, input service.EchoInput, options nexus.StartOperationOptions) (service.EchoOutput, error) {
	// The method is provided with an SDK client that can be used for arbitrary calls such as signaling, querying,
	// and listing workflows but implementations are free to make arbitrary calls to other services or databases, or
	// perform simple computations such as this one.
	return service.EchoOutput(input), nil
})

// Use the NewWorkflowRunOperation constructor, which is the easiest way to expose a workflow as an operation.
// See alternatives at https://pkg.go.dev/go.temporal.io/sdk/temporalnexus.
var HelloOperation = temporalnexus.NewWorkflowRunOperation(service.HelloOperationName, HelloHandlerWorkflow, func(ctx context.Context, input service.HelloInput, options nexus.StartOperationOptions) (client.StartWorkflowOptions, error) {
	return client.StartWorkflowOptions{
		ID: "hello-" + input.Name,
		// Attach to existing workflow if it's already running.
		WorkflowIDConflictPolicy: enumspb.WORKFLOW_ID_CONFLICT_POLICY_USE_EXISTING,
		// Task queue defaults to the task queue this operation is handled on.
	}, nil
})

func HelloHandlerWorkflow(ctx workflow.Context, input service.HelloInput) (service.HelloOutput, error) {
	if err := workflow.Sleep(ctx, 30*time.Second); err != nil {
		return service.HelloOutput{}, nil
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

// @@@SNIPEND

