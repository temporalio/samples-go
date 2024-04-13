package asyncactivity

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
)

func Test_Workflow(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()
	env.RegisterActivity(HelloActivity)
	env.RegisterActivity(ByeActivity)

	// Mock the activities to skip the timers (and avoid test timeout).
	env.OnActivity(HelloActivity, mock.Anything, "Temporal").Return("Hello Temporal!", nil)
	env.OnActivity(ByeActivity, mock.Anything, "Temporal").Return("Bye Temporal!", nil)

	env.ExecuteWorkflow(AsyncActivityWorkflow, "Temporal")

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())

	var result string
	require.NoError(t, env.GetWorkflowResult(&result))
	require.Equal(t, "Hello Temporal!\nBye Temporal!", result)
}
