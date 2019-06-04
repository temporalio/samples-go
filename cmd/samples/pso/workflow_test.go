package main

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/encoded"
	"go.uber.org/cadence/testsuite"
)

func Test_Workflow(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	var activityCalled []string

	//env.OnActivity(evaluateFitnessActivity, mock.Anything, "sphere", []float64{1.0, 2.0, 3.0}).Return(14.0, nil)
	env.SetOnActivityStartedListener(func(activityInfo *activity.Info, ctx context.Context, args encoded.Values) {
		activityType := activityInfo.ActivityType.Name
		activityCalled = append(activityCalled, activityType)
		switch activityType {
		case "initParticleActivityName":
			// var input string
			// s.NoError(args.Get(&input))
			// s.Equal(fileID, input)
		case "updateParticleActivityName":
			// var input string
			// s.NoError(args.Get(&input))
			// s.Equal(fileID, input)
		default:
			panic("unexpected activity call")
		}
	})

	env.ExecuteWorkflow(PSOWorkflow, "sphere")

	//env.AssertExpectations(t) // assert any OnActivity and Return
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	//require.Equal(t, "evaluateFitnessActivity", activityCalled) //activityCalled is a vector with many activities called
}
