package snappycompress

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/testsuite"
)

func Test_Workflow(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	// Mock activity implementation
	env.OnActivity(Activity, mock.Anything, mock.Anything).Return("Hello Temporal!", nil)

	env.ExecuteWorkflow(Workflow, "Temporal")

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	var result string
	require.NoError(t, env.GetWorkflowResult(&result))
	require.Equal(t, "Hello Temporal!", result)
}

func Test_DataConverter(t *testing.T) {
	defConv := converter.GetDefaultDataConverter()
	snappyConv := AlwaysCompressDataConverter

	defaultPayloads, err := defConv.ToPayloads("Testing")
	require.NoError(t, err)

	compressedPayloads, err := snappyConv.ToPayloads("Testing")
	require.NoError(t, err)

	require.NotEqual(t, defaultPayloads.Payloads[0].GetData(), compressedPayloads.Payloads[0].GetData())

	var result string
	err = snappyConv.FromPayloads(compressedPayloads, &result)
	require.NoError(t, err)

	require.Equal(t, "Testing", result)
}
