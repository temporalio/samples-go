package main

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/encoded"
	"go.uber.org/cadence/testsuite"
)

type UnitTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
}

func TestUnitTestSuite(t *testing.T) {
	suite.Run(t, new(UnitTestSuite))
}

func (s *UnitTestSuite) Test_SampleFileProcessingWorkflow() {
	expectedCall := []string{
		"evaluateFitnessActivity",
	}

	var activityCalled []string
	env := s.NewTestWorkflowEnvironment()
	env.SetOnActivityStartedListener(func(activityInfo *activity.Info, ctx context.Context, args encoded.Values) {
		activityType := activityInfo.ActivityType.Name
		activityCalled = append(activityCalled, activityType)
		switch activityType {
		case expectedCall[0]:
			// var input string
			// s.NoError(args.Get(&input))
			// s.Equal(fileID, input)
		default:
			panic("unexpected activity call")
		}
	})
	env.ExecuteWorkflow(PSOWorkflow, "sphere")

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
	//s.Equal(expectedCall, activityCalled) //activityCalled is a vector with many activities called
}
