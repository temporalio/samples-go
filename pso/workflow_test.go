package pso

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
)

func Test_Workflow(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	env.RegisterWorkflow(PSOChildWorkflow)

	env.RegisterActivityWithOptions(
		InitParticleActivity,
		activity.RegisterOptions{Name: InitParticleActivityName},
	)
	env.RegisterActivityWithOptions(
		UpdateParticleActivity,
		activity.RegisterOptions{Name: UpdateParticleActivityName},
	)

	var activityCalled []string

	var dataConverter = NewJSONDataConverter()
	env.SetDataConverter(dataConverter)

	// env.SetWorkflowTimeout(5 * time.Minute)
	// env.SetTestTimeout(5 * time.Minute)

	env.SetOnActivityStartedListener(func(activityInfo *activity.Info, ctx context.Context, args converter.EncodedValues) {
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
	env.SetOnChildWorkflowStartedListener(func(workflowInfo *workflow.Info, ctx workflow.Context, args converter.EncodedValues) {
		childWorkflowID = workflowInfo.WorkflowExecution.ID
	})

	env.ExecuteWorkflow(PSOWorkflow, "sphere")

	require.True(t, env.IsWorkflowCompleted())
	queryAndVerify(t, env, "child", childWorkflowID)
	//queryAndVerify(t, env, "iteration", "???")
	// consider recreating a new test env on every iteration and calling execute workflow
	// with the arguments from the previous iteration (contained in ContinueAsNewError)
	err := env.GetWorkflowError()
	var continueAsNewErr *workflow.ContinueAsNewError
	require.True(t, errors.As(err, &continueAsNewErr))
	require.Equal(t, "continue as new", continueAsNewErr.Error())
}

func queryAndVerify(t *testing.T, env *testsuite.TestWorkflowEnvironment, query string, expectedState string) {
	result, err := env.QueryWorkflow(query)
	require.NoError(t, err)
	var state string
	err = result.Get(&state)
	require.NoError(t, err)
	require.Equal(t, expectedState, state)
}
