package helloworld

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.temporal.io/temporal/activity"
	"go.temporal.io/temporal/encoded"
	"go.temporal.io/temporal/testsuite"
)

func Test_Workflow(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()
	env.RegisterActivity(HelloworldActivity)
	var activityMessage string
	env.SetOnActivityCompletedListener(func(activityInfo *activity.Info, result encoded.Value, err error) {
		_ = result.Get(&activityMessage)
	})
	env.ExecuteWorkflow(HelloworldWorkflow, "world")
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	require.Equal(t, "Hello world!", activityMessage)
}
