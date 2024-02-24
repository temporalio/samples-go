package typedsearchattributes

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/testsuite"
)

func Test_Workflow(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	// mock search attributes on start
	_ = env.SetTypedSearchAttributesOnStart(temporal.NewSearchAttributes(CustomIntKey.ValueSet(1)))

	// mock upsert operations
	env.OnUpsertTypedSearchAttributes(
		temporal.NewSearchAttributes(
			CustomIntKey.ValueSet(2),
			CustomKeyword.ValueSet("Update1"),
			CustomBool.ValueSet(true),
			CustomDouble.ValueSet(3.14),
			CustomDatetimeField.ValueSet(env.Now().UTC()),
			CustomStringField.ValueSet("String field is for text. When query, it will be tokenized for partial match."),
		)).Return(nil).Once()

	env.OnUpsertTypedSearchAttributes(
		temporal.NewSearchAttributes(
			CustomKeyword.ValueSet("Update2"),
			CustomIntKey.ValueUnset(),
		)).Return(nil).Once()

	env.ExecuteWorkflow(SearchAttributesWorkflow)
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
}
