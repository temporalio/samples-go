package cryptconverter

import (
	"context"
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
	defaultDc := converter.GetDefaultDataConverter()

	ctx := context.Background()
	ctx = context.WithValue(ctx, PropagateKey, CryptContext{KeyId: "test"})

	cryptDc := NewCryptDataConverter(
		converter.GetDefaultDataConverter(),
	)
	cryptDcWc := cryptDc.WithContext(ctx)

	defaultPayloads, err := defaultDc.ToPayloads("Testing")
	require.NoError(t, err)

	encryptedPayloads, err := cryptDcWc.ToPayloads("Testing")
	require.NoError(t, err)

	require.NotEqual(t, defaultPayloads.Payloads[0].GetData(), encryptedPayloads.Payloads[0].GetData())

	var result string
	err = cryptDc.FromPayloads(encryptedPayloads, &result)
	require.NoError(t, err)

	require.Equal(t, "Testing", result)
}
