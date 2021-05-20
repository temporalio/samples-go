package cryptconverter

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
	env.OnActivity(Activity, mock.Anything, "Temporal").Return("Hello Temporal!", nil)

	env.ExecuteWorkflow(Workflow, "Temporal")

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	var result string
	require.NoError(t, env.GetWorkflowResult(&result))
	require.Equal(t, "Hello Temporal!", result)
}

func Test_DataConverter(t *testing.T) {
	defaultDc := converter.GetDefaultDataConverter()

	cryptDc := NewCryptDataConverter(
		converter.GetDefaultDataConverter(),
	)

	defaultPayload, err := defaultDc.ToPayload("Testing")
	require.NoError(t, err)

	encryptedPayload, err := cryptDc.ToPayload("Testing")
	require.NoError(t, err)

	require.NotEqual(t, defaultPayload.GetData(), encryptedPayload.GetData())

	var result string
	err = cryptDc.FromPayload(encryptedPayload, &result)
	require.NoError(t, err)

	require.Equal(t, "Testing", result)
}
