package greetings

import (
	"testing"

	"github.com/stretchr/testify/suite"
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
	env := s.NewTestWorkflowEnvironment()
	name := "World"
	greeting := "Hello"
	env.On("GetNameActivity", ).Return(name)
	env.On("GetGreeting", ).Return(greeting)
	env.On("SayGreeting", "Hello", "World").Return(name + " " + greeting + "!")

	env.ExecuteWorkflow(GreetingSample)

	env.AssertCalled(s.T(), "SayGreeting", "Hello", "World")

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
}
