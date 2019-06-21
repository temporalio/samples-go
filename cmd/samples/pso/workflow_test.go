package main

import (
	"context"
	"testing"

	"go.uber.org/cadence/workflow"

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

	var dataConverter = NewGobDataConverter()
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

	var childWorkflowID string
	env.SetOnChildWorkflowStartedListener(func(workflowInfo *workflow.Info, ctx workflow.Context, args encoded.Values) {
		childWorkflowID = workflowInfo.WorkflowExecution.ID
	})

	env.ExecuteWorkflow(PSOWorkflow, "sphere")

	require.True(t, env.IsWorkflowCompleted())
	queryAndVerify(t, env, "child", childWorkflowID)
	require.Equal(t, env.GetWorkflowError().Error(), "ContinueAsNew") // consider recreating a new test env on every iteration and calling execute workflow with the arguments from the previous iteration (contained in ContinueAsNewError)
}

func queryAndVerify(t *testing.T, env *testsuite.TestWorkflowEnvironment, query string, expectedState string) {
	result, err := env.QueryWorkflow(query)
	require.NoError(t, err)
	var state string
	err = result.Get(&state)
	require.NoError(t, err)
	require.Equal(t, expectedState, state)
}
