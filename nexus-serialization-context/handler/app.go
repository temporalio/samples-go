package handler

import (
	"context"
	"fmt"

	"github.com/nexus-rpc/sdk-go/nexus"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporalnexus"
	"go.temporal.io/sdk/workflow"

	"github.com/temporalio/samples-go/nexus-serialization-context/service"
)

var SyncOperation = nexus.NewSyncOperation(service.SyncOperationName,
	func(_ context.Context, input service.EchoInput, _ nexus.StartOperationOptions) (service.EchoOutput, error) {
		return service.EchoOutput(input), nil
	})

var AsyncOperation = temporalnexus.NewWorkflowRunOperation(service.AsyncOperationName, EchoWorkflow,
	func(_ context.Context, _ service.EchoInput, options nexus.StartOperationOptions) (client.StartWorkflowOptions, error) {
		// Retries reuse the Nexus request ID. Independent operations get separate workflows.
		if options.RequestID == "" {
			return client.StartWorkflowOptions{}, fmt.Errorf("missing Nexus request ID")
		}
		return client.StartWorkflowOptions{ID: "nexus-echo-" + options.RequestID}, nil
	})

func EchoWorkflow(_ workflow.Context, input service.EchoInput) (service.EchoOutput, error) {
	return service.EchoOutput(input), nil
}
