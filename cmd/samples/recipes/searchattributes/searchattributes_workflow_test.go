package main

import (
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/temporal/.gen/go/shared"
	"go.temporal.io/temporal/testsuite"
	"testing"
	"time"
)

func Test_Workflow(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	// mock search attributes on start
	env.SetSearchAttributesOnStart(getSearchAttributesForStart())

	// mock upsert operations
	attributes := map[string]interface{}{
		"CustomIntField":      2, // update CustomIntField from 1 to 2, then insert other fields
		"CustomKeywordField":  "Update1",
		"CustomBoolField":     true,
		"CustomDoubleField":   3.14,
		"CustomDatetimeField": time.Date(2019, 1, 1, 0, 0, 0, 0, time.Local),
		"CustomStringField":   "String field is for text. When query, it will be tokenized for partial match. StringTypeField cannot be used in Order By",
	}
	env.OnUpsertSearchAttributes(attributes).Return(nil).Once()

	attributes = map[string]interface{}{
		"CustomKeywordField": "Update2",
	}
	env.OnUpsertSearchAttributes(attributes).Return(nil).Once()

	// mock activity
	env.OnActivity(listExecutions, mock.Anything, mock.Anything).Return([]*shared.WorkflowExecutionInfo{{}}, nil).Once()

	env.ExecuteWorkflow(SearchAttributesWorkflow)
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
}
