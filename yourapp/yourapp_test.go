// @@@SNIPSTART samples-go-yourapp-tests
package yourapp

import (
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
)

func Test_Workflow(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	// Mock activity implementation
	var activities *YourActivityObject
	activityResult := YourActivityResultObject{
		ResultFieldX: "Hello World!",
		ResultFieldY: 1,
	}
	env.OnActivity(activities.YourActivityDefinition, mock.Anything, mock.Anything).Return(activityResult, nil)
	env.OnActivity(activities.PrintSharedSate, mock.Anything).Return(nil)
	env.ExecuteWorkflow(YourWorkflowDefinition, YourWorkflowParam{})

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	var result YourWorkflowResultObject
	require.NoError(t, env.GetWorkflowResult(&result))
}

func Test_Activity(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestActivityEnvironment()

	var activities YourActivityObject

	message := "No messages!"
	counter := 0
	activities.SharedMessageState = &message
	activities.SharedCounterState = &counter

	env.RegisterActivity(activities.YourActivityDefinition)

	activityParam := YourActivityParam{
		ActivityParamX: "Hello",
		ActivityParamY: 0,
	}
	val, err := env.ExecuteActivity(activities.YourActivityDefinition, activityParam)
	require.NoError(t, err)
	var res YourActivityResultObject
	require.NoError(t, val.Get(&res))
	require.Equal(t, "Hello World!", res.ResultFieldX)
}
// @@@SNIPEND
