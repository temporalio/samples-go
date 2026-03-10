package handler

import (
	"context"
	"fmt"

	"github.com/nexus-rpc/sdk-go/nexus"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporalnexus"
	"go.temporal.io/sdk/workflow"

	"github.com/temporalio/samples-go/nexus/service"
)

// If you want to map a workflow with multiple arguments to a nexus operation use NewWorkflowRunOperationWithOptions or MustNewWorkflowRunOperationWithOptions.
// See alternatives at https://pkg.go.dev/go.temporal.io/sdk/temporalnexus.
// @@@SNIPSTART  samples-go-nexus-handler-multiargs
var HelloOperation = temporalnexus.MustNewWorkflowRunOperationWithOptions(temporalnexus.WorkflowRunOperationOptions[service.HelloInput, service.HelloOutput]{
	Name: service.HelloOperationName,
	Handler: func(ctx context.Context, input service.HelloInput, options nexus.StartOperationOptions) (temporalnexus.WorkflowHandle[service.HelloOutput], error) {
		return temporalnexus.ExecuteUntypedWorkflow[service.HelloOutput](
			ctx,
			options,
			client.StartWorkflowOptions{
				// Workflow IDs should typically be business meaningful IDs and are used to dedupe workflow starts.
				// For this example, we're using the request ID allocated by Temporal when the caller workflow schedules
				// the operation, this ID is guaranteed to be stable across retries of this operation.
				ID: options.RequestID,
			},
			HelloHandlerWorkflow,
			input.Name,
			input.Language,
		)
	},
})

// @@@SNIPEND

func HelloHandlerWorkflow(ctx workflow.Context, name string, language service.Language) (service.HelloOutput, error) {
	switch language {
	case service.EN:
		return service.HelloOutput{Message: "Hello " + name + " 👋"}, nil
	case service.FR:
		return service.HelloOutput{Message: "Bonjour " + name + " 👋"}, nil
	case service.DE:
		return service.HelloOutput{Message: "Hallo " + name + " 👋"}, nil
	case service.ES:
		return service.HelloOutput{Message: "¡Hola! " + name + " 👋"}, nil
	case service.TR:
		return service.HelloOutput{Message: "Merhaba " + name + " 👋"}, nil
	}
	return service.HelloOutput{}, fmt.Errorf("unsupported language %q", language)
}
