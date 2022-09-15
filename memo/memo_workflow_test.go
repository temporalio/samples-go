package memo

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
	env.RegisterActivity(DescribeWorkflow)

	// mock search attributes on start
	_ = env.SetMemoOnStart(map[string]interface{}{"description": "Test memo workflow"})

	// mock upsert operations
	memo := map[string]interface{}{
		"description": "Test upsert memo workflow",
	}
	env.OnUpsertMemo(memo).Return(nil).Once()

	// mock activity
	env.OnActivity(DescribeWorkflow, mock.Anything, mock.Anything).Return(&workflowpb.WorkflowExecutionInfo{}, nil).Once()

	env.ExecuteWorkflow(MemoWorkflow)
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
}
