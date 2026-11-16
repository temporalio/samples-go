package caller

import (
	"go.temporal.io/sdk/workflow"

	"github.com/temporalio/samples-go/nexus-per-endpoint-encryption/service"
)

type operationPair struct{ sync, async workflow.NexusOperationFuture }

// EncryptionWorkflow starts every operation before waiting, so different endpoint keys are in flight together.
func EncryptionWorkflow(ctx workflow.Context, message string) (service.WorkflowOutput, error) {
	compressed1 := startOperations(ctx, service.CompressedEncrypted1Endpoint, message)
	compressed2 := startOperations(ctx, service.CompressedEncrypted2Endpoint, message)
	encrypted1 := startOperations(ctx, service.Encrypted1Endpoint, message)
	encrypted2 := startOperations(ctx, service.Encrypted2Endpoint, message)
	var result service.WorkflowOutput
	var err error
	if result.CompressedEncrypted1, err = compressed1.results(ctx); err != nil {
		return result, err
	}
	if result.CompressedEncrypted2, err = compressed2.results(ctx); err != nil {
		return result, err
	}
	if result.Encrypted1, err = encrypted1.results(ctx); err != nil {
		return result, err
	}
	if result.Encrypted2, err = encrypted2.results(ctx); err != nil {
		return result, err
	}
	return result, nil
}

func startOperations(ctx workflow.Context, endpoint, message string) operationPair {
	nexusClient := workflow.NewNexusClient(endpoint, service.Name)
	input := service.EchoInput{Message: message}
	return operationPair{
		sync:  nexusClient.ExecuteOperation(ctx, service.SyncOperationName, input, workflow.NexusOperationOptions{}),
		async: nexusClient.ExecuteOperation(ctx, service.AsyncOperationName, input, workflow.NexusOperationOptions{}),
	}
}

func (p operationPair) results(ctx workflow.Context) (service.EndpointResults, error) {
	var syncResult, asyncResult service.EchoOutput
	if err := p.sync.Get(ctx, &syncResult); err != nil {
		return service.EndpointResults{}, err
	}
	if err := p.async.Get(ctx, &asyncResult); err != nil {
		return service.EndpointResults{}, err
	}
	return service.EndpointResults{Sync: syncResult.Message, Async: asyncResult.Message}, nil
}
