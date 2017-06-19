package main

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/cadence"
)

type UnitTestSuite struct {
	suite.Suite
	cadence.WorkflowTestSuite
}

func TestUnitTestSuite(t *testing.T) {
	suite.Run(t, new(UnitTestSuite))
}

func (s *UnitTestSuite) Test_SampleGreetingsWorkflow() {
	sayGreetingActivityName := "github.com/samarabbas/cadence-samples/cmd/samples/recipes/greetings.sayGreetingActivity"
	env := s.NewTestWorkflowEnvironment()
	var startCalled, endCalled bool
	env.SetOnActivityStartedListener(func(activityInfo *cadence.ActivityInfo, ctx context.Context, args cadence.EncodedValues) {
		if sayGreetingActivityName == activityInfo.ActivityType.Name {
			var greeting, name string
			args.Get(&greeting, &name)
			s.Equal("Hello", greeting)
			s.Equal("Cadence", name)
			startCalled = true
		}
	})
	env.SetOnActivityCompletedListener(func(activityInfo *cadence.ActivityInfo, result cadence.EncodedValue, err error) {
		if sayGreetingActivityName == activityInfo.ActivityType.Name {
			var sayResult string
			result.Get(&sayResult)
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
