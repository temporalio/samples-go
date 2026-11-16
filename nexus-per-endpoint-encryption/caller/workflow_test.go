package caller_test

import (
	"testing"

	"github.com/nexus-rpc/sdk-go/nexus"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/testsuite"

	"github.com/temporalio/samples-go/nexus-per-endpoint-encryption/caller"
	"github.com/temporalio/samples-go/nexus-per-endpoint-encryption/handler"
	"github.com/temporalio/samples-go/nexus-per-endpoint-encryption/service"
)

// This checks caller orchestration; native context propagation is verified separately against a server.
func TestEncryptionWorkflow(t *testing.T) {
	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestWorkflowEnvironment()
	env.RegisterWorkflow(caller.EncryptionWorkflow)
	env.RegisterWorkflow(handler.EchoWorkflow)
	svc := nexus.NewService(service.Name)
	require.NoError(t, svc.Register(handler.SyncOperation, handler.AsyncOperation))
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
	require.Equal(t, map[string]int{service.SyncOperationName: 4, service.AsyncOperationName: 4}, counts)
	var result service.WorkflowOutput
	require.NoError(t, env.GetWorkflowResult(&result))
	expected := service.EndpointResults{Sync: "Hello from Nexus", Async: "Hello from Nexus"}
	require.Equal(t, service.WorkflowOutput{
		CompressedEncrypted1: expected, CompressedEncrypted2: expected,
		Encrypted1: expected, Encrypted2: expected,
	}, result)
}
