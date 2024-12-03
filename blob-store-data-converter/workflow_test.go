package blobstore_data_converter

import (
	"github.com/stretchr/testify/mock"
	commonpb "go.temporal.io/api/common/v1"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/workflow"
	"testing"

	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
)

func Test_Workflow(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	// Set up the environment with the expected context propagators and data converter
	env.SetContextPropagators([]workflow.ContextPropagator{NewContextPropagator()})
	headerDC := converter.GetDefaultDataConverter()
	p, err := headerDC.ToPayload(PropagatedValues{
		TenantID:       "test-tenant",
		BlobNamePrefix: []string{t.Name()},
	})
	require.NoError(t, err)
	env.SetHeader(&commonpb.Header{
		Fields: map[string]*commonpb.Payload{
			propagationKey: p,
		},
	})

	// Mock activity implementation
	env.OnActivity(Activity, mock.Anything, mock.Anything).Return("Hello From TestActivity!", nil)

	env.ExecuteWorkflow(Workflow, "Temporal")

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	var result string
	require.NoError(t, env.GetWorkflowResult(&result))
	require.Equal(t, "WorkflowSays: Hello From TestActivity!", result)
}
