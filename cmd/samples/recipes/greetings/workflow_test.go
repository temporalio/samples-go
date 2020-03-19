package main

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.temporal.io/temporal/activity"
	"go.temporal.io/temporal/encoded"
	"go.temporal.io/temporal/testsuite"
)

type UnitTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
}

func TestUnitTestSuite(t *testing.T) {
	suite.Run(t, new(UnitTestSuite))
}

func (s *UnitTestSuite) Test_SampleGreetingsWorkflow() {
	sayGreetingActivityName := "sayGreetingActivity"
	env := s.NewTestWorkflowEnvironment()
	env.RegisterActivity(getGreetingActivity)
	env.RegisterActivity(getNameActivity)
	env.RegisterActivity(sayGreetingActivity)

	var startCalled, endCalled bool
	env.SetOnActivityStartedListener(func(activityInfo *activity.Info, ctx context.Context, args encoded.Values) {
		if sayGreetingActivityName == activityInfo.ActivityType.Name {
			var greeting, name string
			_ = args.Get(&greeting, &name)
			s.Equal("Hello", greeting)
			s.Equal("Cadence", name)
			startCalled = true
		}
	})
	env.SetOnActivityCompletedListener(func(activityInfo *activity.Info, result encoded.Value, err error) {
		if sayGreetingActivityName == activityInfo.ActivityType.Name {
			var sayResult string
			_ = result.Get(&sayResult)
			s.Equal("Greeting: Hello Cadence!\n", sayResult)
			endCalled = true
		}
	})

	env.ExecuteWorkflow(SampleGreetingsWorkflow)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
	s.True(startCalled)
	s.True(endCalled)
}
