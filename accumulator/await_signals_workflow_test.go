package accumulator

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/testsuite"
)

type UnitTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
}

func TestUnitTestSuite(t *testing.T) {
	suite.Run(t, new(UnitTestSuite))
}

func (s *UnitTestSuite) Test_WorkflowTimeout() {
	env := s.NewTestWorkflowEnvironment()
	env.ExecuteWorkflow(AccumulateSignalsWorkflow)

	s.True(env.IsWorkflowCompleted())
	// Workflow times out
	s.Error(env.GetWorkflowError())
}

func (s *UnitTestSuite) Test_SignalsInOrder() {
	env := s.NewTestWorkflowEnvironment()
	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow("Signal1", nil)
	}, time.Hour)
	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow("Signal2", nil)
	}, time.Hour+time.Second)
	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow("Signal3", nil)
	}, time.Hour+3*time.Second)
	env.ExecuteWorkflow(AccumulateSignalsWorkflow)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
}

func (s *UnitTestSuite) Test_SignalsInReverseOrder() {
	env := s.NewTestWorkflowEnvironment()
	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow("Signal3", nil)
	}, time.Hour)
	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow("Signal2", nil)
	}, time.Hour+time.Second)
	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow("Signal1", nil)
	}, time.Hour+3*time.Second)
	env.ExecuteWorkflow(AccumulateSignalsWorkflow)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
}
