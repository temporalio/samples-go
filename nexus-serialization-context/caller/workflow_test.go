package caller_test

import (
	"context"
	"testing"

	"github.com/nexus-rpc/sdk-go/nexus"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/temporalnexus"
	"go.temporal.io/sdk/testsuite"

	"github.com/temporalio/samples-go/nexus-serialization-context/caller"
	"github.com/temporalio/samples-go/nexus-serialization-context/handler"
	"github.com/temporalio/samples-go/nexus-serialization-context/service"
)

// This checks caller orchestration; native context propagation is verified separately against a server.
func TestEncryptionWorkflow(t *testing.T) {
	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestWorkflowEnvironment()
	env.RegisterWorkflow(caller.EncryptionWorkflow)
	env.RegisterWorkflow(handler.EchoWorkflow)
	svc := nexus.NewService(service.Name)
	require.NoError(t, svc.Register(
		&taggedEchoOperation{Operation: handler.SyncOperation},
		&taggedEchoOperation{Operation: handler.AsyncOperation},
	))
	env.RegisterNexusService(svc)
	counts := map[string]int{}
	env.SetOnNexusOperationStartedListener(func(name, operation string, input converter.EncodedValue) {
		require.Equal(t, service.Name, name)
		var decoded service.EchoInput
		require.NoError(t, input.Get(&decoded))
		require.Equal(t, "Hello from Nexus", decoded.Message)
		counts[operation]++
	})
	env.ExecuteWorkflow(caller.EncryptionWorkflow, "Hello from Nexus")
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	// The totals and distinct replies below check that every endpoint/operation pair ran once.
	require.Equal(t, map[string]int{service.SyncOperationName: 4, service.AsyncOperationName: 4}, counts)
	var result service.WorkflowOutput
	require.NoError(t, env.GetWorkflowResult(&result))
	expected := func(endpoint string) service.EndpointResults {
		return service.EndpointResults{
			Sync:  endpoint + "/" + service.SyncOperationName + ": Hello from Nexus",
			Async: endpoint + "/" + service.AsyncOperationName + ": Hello from Nexus",
		}
	}
	require.Equal(t, service.WorkflowOutput{
		CompressedEncrypted1: expected(service.CompressedEncrypted1Endpoint),
		CompressedEncrypted2: expected(service.CompressedEncrypted2Endpoint),
		Encrypted1:           expected(service.Encrypted1Endpoint),
		Encrypted2:           expected(service.Encrypted2Endpoint),
	}, result)
}

// Tag each echo input with its destination so the real handlers return distinct replies.
// Async operations still complete through their backing workflow.
type taggedEchoOperation struct {
	nexus.Operation[service.EchoInput, service.EchoOutput]
}

func (o *taggedEchoOperation) Start(ctx context.Context, input service.EchoInput, options nexus.StartOperationOptions) (nexus.HandlerStartOperationResult[service.EchoOutput], error) {
	endpoint := temporalnexus.GetOperationInfo(ctx).Endpoint
	input.Message = endpoint + "/" + o.Name() + ": " + input.Message
	return o.Operation.Start(ctx, input, options)
}
