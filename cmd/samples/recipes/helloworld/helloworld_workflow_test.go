package main

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/cadence"
)

func Test_Workflow(t *testing.T) {
	testSuite := &cadence.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()
	var activityMessage string
	env.SetOnActivityCompletedListener(func(activityInfo *cadence.ActivityInfo, result cadence.EncodedValue, err error) {
		result.Get(&activityMessage)
	})
	env.ExecuteWorkflow(Workflow, "world")
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	require.Equal(t, "Hello world!", activityMessage)
}
