package searchattributes

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	workflowpb "go.temporal.io/api/workflow/v1"
	"go.temporal.io/sdk/testsuite"
)

func Test_Workflow(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()
	env.RegisterActivity(ListExecutions)

	// mock search attributes on start
	//lint:ignore SA1019 typed alternative available, leaving this for reference
	_ = env.SetSearchAttributesOnStart(map[string]interface{}{"CustomIntField": 1})

	// mock upsert operations
	attributes := map[string]interface{}{
		"CustomIntField":      2, // update CustomIntField from 1 to 2, then insert other fields
		"CustomKeywordField":  "Update1",
		"CustomBoolField":     true,
		"CustomDoubleField":   3.14,
		"CustomDatetimeField": env.Now().UTC(),
		"CustomStringField":   "String field is for text. When query, it will be tokenized for partial match. StringTypeField cannot be used in Order By",
	}
	//lint:ignore SA1019 typed alternative available, leaving this for reference
	env.OnUpsertSearchAttributes(attributes).Return(nil).Once()

	attributes = map[string]interface{}{
		"CustomKeywordField": "Update2",
	}
	//lint:ignore SA1019 typed alternative available, leaving this for reference
	env.OnUpsertSearchAttributes(attributes).Return(nil).Once()

	// mock activity
	env.OnActivity(ListExecutions, mock.Anything, mock.Anything).Return([]*workflowpb.WorkflowExecutionInfo{{}}, nil).Once()

	env.ExecuteWorkflow(SearchAttributesWorkflow)
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
}
