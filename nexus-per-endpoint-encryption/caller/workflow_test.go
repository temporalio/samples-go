package caller_test

import (
	"testing"

	"github.com/nexus-rpc/sdk-go/nexus"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"

	"github.com/temporalio/samples-go/nexus-per-endpoint-encryption/caller"
	"github.com/temporalio/samples-go/nexus-per-endpoint-encryption/encryption"
	"github.com/temporalio/samples-go/nexus-per-endpoint-encryption/handler"
	"github.com/temporalio/samples-go/nexus-per-endpoint-encryption/service"
)

func TestWorkflowCallsSyncAndAsyncOperationsOnBothEndpoints(t *testing.T) {
	var testSuite testsuite.WorkflowTestSuite
	env := testSuite.NewTestWorkflowEnvironment()
	env.SetDataConverter(encryption.NewDataConverter())
	env.RegisterWorkflow(caller.PerEndpointEncryptionWorkflow)
	env.RegisterWorkflow(handler.AsyncOperationWorkflow)

	nexusService := nexus.NewService(service.Name)
	require.NoError(t, nexusService.Register(handler.SyncOperation, handler.AsyncOperation))
	env.RegisterNexusService(nexusService)

	env.ExecuteWorkflow(caller.PerEndpointEncryptionWorkflow)

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	var result service.WorkflowOutput
	require.NoError(t, env.GetWorkflowResult(&result))
	require.Equal(t, service.WorkflowOutput{
		EndpointASync:  "sync: endpoint A",
		EndpointAAsync: "async: endpoint A",
		EndpointBSync:  "sync: endpoint B",
		EndpointBAsync: "async: endpoint B",
	}, result)
}
