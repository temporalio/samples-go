package main

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/encoded"
	"go.uber.org/cadence/testsuite"
	"go.uber.org/cadence/worker"
)

func Test_Workflow(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	var activityCalled []string

	var dataConverter = newGobDataConverter()
	workerOptions := worker.Options{
		DataConverter: dataConverter,
	}
	env.SetWorkerOptions(workerOptions)

	// env.SetWorkflowTimeout(time.Minute * 5)
	// env.SetTestTimeout(time.Minute * 5)

	env.SetOnActivityStartedListener(func(activityInfo *activity.Info, ctx context.Context, args encoded.Values) {
		activityType := activityInfo.ActivityType.Name
		activityCalled = append(activityCalled, activityType)
		switch activityType {
		case "initParticleActivityName":
		case "updateParticleActivityName":
		default:
			panic("unexpected activity call")
		}
	})

	env.ExecuteWorkflow(PSOWorkflow, "sphere")

	//env.AssertExpectations(t) // assert any OnActivity and Return
	require.True(t, env.IsWorkflowCompleted())
	//require.NoError(t, env.GetWorkflowError())
	require.Equal(t, env.GetWorkflowError().Error(), "ContinueAsNew") // consider recreating a new test env on every iteration and calling execute workflow with the arguments from the previous iteration (contained in ContinueAsNewError)
	//require.Equal(t, "evaluateFitnessActivity", activityCalled) //activityCalled is a vector with many activities called
	queryAndVerify(t, env, "ContinueAsNew issued")
}

func queryAndVerify(t *testing.T, env *testsuite.TestWorkflowEnvironment, expectedState string) {
	result, err := env.QueryWorkflow("state")
	require.NoError(t, err)
	var state string
	err = result.Get(&state)
	require.NoError(t, err)
	require.Equal(t, expectedState, state)
}
