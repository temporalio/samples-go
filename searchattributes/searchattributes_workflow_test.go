package searchattributes

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	commonpb "go.temporal.io/api/common/v1"
	workflowpb "go.temporal.io/api/workflow/v1"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/testsuite"
)

func Test_Workflow(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	// mock search attributes on start
	_ = env.SetTypedSearchAttributesOnStart(temporal.NewSearchAttributes(CustomIntField.ValueSet(1)))

	// mock upsert operations
	env.OnUpsertTypedSearchAttributes(
		temporal.NewSearchAttributes(
			CustomIntField.ValueSet(2), // update CustomIntField from 1 to 2, then insert other fields
			CustomKeywordField.ValueSet("Keyword fields supports prefix search"),
			CustomBoolField.ValueSet(true),
			CustomDoubleField.ValueSet(3.14),
			CustomDatetimeField.ValueSet(env.Now().UTC()),
			CustomKeywordListField.ValueSet([]string{"value1", "value2"}),
		)).Return(nil).Once()

	env.OnUpsertTypedSearchAttributes(
		temporal.NewSearchAttributes(
			CustomDoubleField.ValueUnset(),
		)).Return(nil).Once()

	env.OnActivity(ListExecutions, mock.Anything, mock.Anything).Return(
		[]*workflowpb.WorkflowExecutionInfo{
			{
				Execution: &commonpb.WorkflowExecution{},
			},
		},
		nil,
	).Once()

	env.ExecuteWorkflow(SearchAttributesWorkflow)
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
}
