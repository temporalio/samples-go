package activities_async

import (
	"context"
	"errors"
	"github.com/stretchr/testify/mock"
	"testing"

	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
)

func Test_Workflow_Succeeds(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	env.RegisterActivity(SayHello)
	env.RegisterActivity(SayGoodbye)
	env.ExecuteWorkflow(AsyncActivitiesWorkflow)

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	var res string
	err := env.GetWorkflowResult(&res)
	require.NoError(t, err)
	require.Equal(t, "Hello Temporal! It was great to meet you, but time has come. Goodbye Temporal!", res)
}

func testWorkflowFailsWhenActivityFail(t *testing.T, a interface{}) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()
	env.RegisterActivity(SayHello)
	env.RegisterActivity(SayGoodbye)
	env.OnActivity(a, mock.Anything, mock.Anything).Return(func(ctx context.Context, arg string) (string, error) {
		return "", errors.New("activity failed")
	})
	env.ExecuteWorkflow(AsyncActivitiesWorkflow)
	require.True(t, env.IsWorkflowCompleted())
	require.Error(t, env.GetWorkflowError())
}

func Test_Workflow_FailsIfSayHelloFails(t *testing.T) {
	testWorkflowFailsWhenActivityFail(t, SayHello)
}
func Test_Workflow_FailsIfSayGoodbyeFails(t *testing.T) {
	testWorkflowFailsWhenActivityFail(t, SayGoodbye)
}
