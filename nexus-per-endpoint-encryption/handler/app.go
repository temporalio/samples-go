package handler

import (
	"context"

	"github.com/nexus-rpc/sdk-go/nexus"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporalnexus"
	"go.temporal.io/sdk/workflow"

	"github.com/temporalio/samples-go/nexus-per-endpoint-encryption/service"
)

var SyncOperation = nexus.NewSyncOperation(
	service.SyncOperationName,
	func(_ context.Context, input service.OperationInput, _ nexus.StartOperationOptions) (service.OperationOutput, error) {
		return service.OperationOutput{Message: "sync: " + input.Message}, nil
	},
)

var AsyncOperation = temporalnexus.NewWorkflowRunOperation(
	service.AsyncOperationName,
	AsyncOperationWorkflow,
	func(_ context.Context, input service.OperationInput, _ nexus.StartOperationOptions) (client.StartWorkflowOptions, error) {
		return client.StartWorkflowOptions{ID: "encrypted-operation-" + input.RequestID}, nil
	},
)

func AsyncOperationWorkflow(_ workflow.Context, input service.OperationInput) (service.OperationOutput, error) {
	return service.OperationOutput{Message: "async: " + input.Message}, nil
}
