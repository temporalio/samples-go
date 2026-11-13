package handler

import (
	"context"
	"fmt"

	"github.com/nexus-rpc/sdk-go/nexus"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporalnexus"
	"go.temporal.io/sdk/workflow"

	"github.com/temporalio/samples-go/ctxpropagation"
	"github.com/temporalio/samples-go/nexus/service"
)

// NewSyncOperation is a meant for exposing simple RPC handlers.
var EchoOperation = nexus.NewSyncOperation(service.EchoOperationName, func(ctx context.Context, input service.EchoInput, options nexus.StartOperationOptions) (service.EchoOutput, error) {
	// Values may be extracted from the context in the Operation handler body.
	values, ok := ctx.Value(ctxpropagation.PropagateKey).(ctxpropagation.Values)
	if ok {
		input.Message += ", " + values.Key + ": " + values.Value
	}
	// Use temporalnexus.GetClient to get the client that the worker was initialized with to perform client calls
	// such as signaling, querying, and listing workflows. Implementations are free to make arbitrary calls to other
	// services or databases, or perform simple computations such as this one.
	return service.EchoOutput(input), nil
})

// Use the NewWorkflowRunOperation constructor, which is the easiest way to expose a workflow as an operation.
// See alternatives at https://pkg.go.dev/go.temporal.io/sdk/temporalnexus.
var HelloOperation = temporalnexus.NewWorkflowRunOperation(service.HelloOperationName, HelloHandlerWorkflow, func(ctx context.Context, input service.HelloInput, options nexus.StartOperationOptions) (client.StartWorkflowOptions, error) {
	// Values may be extracted from the context in the Operation handler body if necessary, this sample propagates
	// the context to the handler workflow.
	// values, ok := ctx.Value(ctxpropagation.PropagateKey).(ctxpropagation.Values)

	return client.StartWorkflowOptions{
		// Workflow IDs should typically be business meaningful IDs and are used to dedupe workflow starts.
		// For this example, use a business ID derived from the greeting input so repeated operations
		// for the same name and language resolve to the same workflow.
		ID: service.HelloWorkflowID(input),
		// Task queue defaults to the task queue this operation is handled on.
	}, nil
})

func HelloHandlerWorkflow(ctx workflow.Context, input service.HelloInput) (service.HelloOutput, error) {
	// Values may be extracted from the handler workflow context.
	values, ok := ctx.Value(ctxpropagation.PropagateKey).(ctxpropagation.Values)
	if ok {
		input.Name += ", " + values.Key + ": " + values.Value
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
