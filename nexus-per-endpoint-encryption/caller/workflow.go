package caller

import (
	"go.temporal.io/sdk/workflow"

	"github.com/temporalio/samples-go/nexus-per-endpoint-encryption/service"
)

const TaskQueue = "nexus-encryption-caller-task-queue"

func PerEndpointEncryptionWorkflow(ctx workflow.Context) (service.WorkflowOutput, error) {
	endpointA := workflow.NewNexusClient(service.EndpointA, service.Name)
	endpointB := workflow.NewNexusClient(service.EndpointB, service.Name)
	workflowID := workflow.GetInfo(ctx).WorkflowExecution.ID

	endpointASync := endpointA.ExecuteOperation(
		ctx,
		service.SyncOperationName,
		service.OperationInput{Message: "endpoint A", RequestID: workflowID + "-endpoint-a-sync"},
		workflow.NexusOperationOptions{},
	)
	endpointAAsync := endpointA.ExecuteOperation(
		ctx,
		service.AsyncOperationName,
		service.OperationInput{Message: "endpoint A", RequestID: workflowID + "-endpoint-a-async"},
		workflow.NexusOperationOptions{},
	)
	endpointBSync := endpointB.ExecuteOperation(
		ctx,
		service.SyncOperationName,
		service.OperationInput{Message: "endpoint B", RequestID: workflowID + "-endpoint-b-sync"},
		workflow.NexusOperationOptions{},
	)
	endpointBAsync := endpointB.ExecuteOperation(
		ctx,
		service.AsyncOperationName,
		service.OperationInput{Message: "endpoint B", RequestID: workflowID + "-endpoint-b-async"},
		workflow.NexusOperationOptions{},
	)

	var endpointASyncResult service.OperationOutput
	if err := endpointASync.Get(ctx, &endpointASyncResult); err != nil {
		return service.WorkflowOutput{}, err
	}
	var endpointAAsyncResult service.OperationOutput
	if err := endpointAAsync.Get(ctx, &endpointAAsyncResult); err != nil {
		return service.WorkflowOutput{}, err
	}
	var endpointBSyncResult service.OperationOutput
	if err := endpointBSync.Get(ctx, &endpointBSyncResult); err != nil {
		return service.WorkflowOutput{}, err
	}
	var endpointBAsyncResult service.OperationOutput
	if err := endpointBAsync.Get(ctx, &endpointBAsyncResult); err != nil {
		return service.WorkflowOutput{}, err
	}
	return service.WorkflowOutput{
		EndpointASync:  endpointASyncResult.Message,
		EndpointAAsync: endpointAAsyncResult.Message,
		EndpointBSync:  endpointBSyncResult.Message,
		EndpointBAsync: endpointBAsyncResult.Message,
	}, nil
}
