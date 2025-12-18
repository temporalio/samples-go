package caller_test

import (
	"context"
	"testing"

	"github.com/nexus-rpc/sdk-go/nexus"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporalnexus"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"

	"github.com/temporalio/samples-go/nexus/caller"
	"github.com/temporalio/samples-go/nexus/service"
)

var EchoOperation = nexus.NewSyncOperation(service.EchoOperationName, func(ctx context.Context, input service.EchoInput, options nexus.StartOperationOptions) (service.EchoOutput, error) {
	// NOTE: temporalnexus.GetClient is not usable in the test environment.
	return service.EchoOutput(input), nil
})

var HelloOperation = temporalnexus.NewWorkflowRunOperation(service.HelloOperationName, FakeHelloHandlerWorkflow, func(ctx context.Context, input service.HelloInput, options nexus.StartOperationOptions) (client.StartWorkflowOptions, error) {
	return client.StartWorkflowOptions{
		// Do not use RequestID for production use cases. ID should be a meaninful business ID.
		ID: options.RequestID,
	}, nil
})

func FakeHelloHandlerWorkflow(_ workflow.Context, input service.HelloInput) (service.HelloOutput, error) {
	return service.HelloOutput{Message: "fake:" + string(input.Language) + ":" + input.Name}, nil
}

func Test_Echo(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	env.RegisterWorkflow(caller.EchoCallerWorkflow)

	s := nexus.NewService(service.HelloServiceName)
	require.NoError(t, s.Register(EchoOperation))
	env.RegisterNexusService(s)

	env.ExecuteWorkflow(caller.EchoCallerWorkflow, "hey")

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	var result string
	require.NoError(t, env.GetWorkflowResult(&result))
	require.Equal(t, "hey", result)
}

func Test_Hello(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	env.RegisterWorkflow(caller.HelloCallerWorkflow)
	env.RegisterWorkflow(FakeHelloHandlerWorkflow)

	s := nexus.NewService(service.HelloServiceName)
	require.NoError(t, s.Register(HelloOperation))
	env.RegisterNexusService(s)

	env.ExecuteWorkflow(caller.HelloCallerWorkflow, "test", service.DE)

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	var result string
	require.NoError(t, env.GetWorkflowResult(&result))
	require.Equal(t, "fake:de:test", result)
}
