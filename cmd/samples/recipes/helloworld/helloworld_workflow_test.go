package main

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/encoded"
	"go.uber.org/cadence/testsuite"
)

func Test_Workflow(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()
	var activityMessage string
	env.SetOnActivityCompletedListener(func(activityInfo *activity.Info, result encoded.Value, err error) {
		result.Get(&activityMessage)
	})
	env.ExecuteWorkflow(Workflow, "world")
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	require.Equal(t, "Hello world!", activityMessage)
}
