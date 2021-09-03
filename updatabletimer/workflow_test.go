package updatabletimer

import (
	"errors"
	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/testsuite"
	"testing"
	"time"
)

type UnitTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
}

func TestUnitTestSuite(t *testing.T) {
	suite.Run(t, new(UnitTestSuite))
}

func (s *UnitTestSuite) Test_Sleep() {
	env := s.NewTestWorkflowEnvironment()
	start := env.Now()
	wakeUpTime := env.Now().Add(30 * time.Minute)
	env.RegisterDelayedCallback(func() {
		value, err := env.QueryWorkflow(QueryType)
		s.NoError(err)
		var queryResult time.Time
		err = value.Get(&queryResult)
		s.NoError(err)
		s.True(wakeUpTime.Equal(queryResult))
	}, time.Minute*10)
	env.ExecuteWorkflow(Workflow, wakeUpTime)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
	elapsed := env.Now().Sub(start)
	s.Equal(elapsed, 30*time.Minute)
}

func (s *UnitTestSuite) Test_UpdateWakeUpTime() {
	env := s.NewTestWorkflowEnvironment()
	start := env.Now()

	// Update wake-up time
	updatedWakeUpTime1 := env.Now().Add(15 * time.Minute)
	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow(SignalType, updatedWakeUpTime1)
	}, time.Minute*10)

	// Check that it was updated
	env.RegisterDelayedCallback(func() {
		value, err := env.QueryWorkflow(QueryType)
		s.NoError(err)
		var queryResult time.Time
		err = value.Get(&queryResult)
		s.NoError(err)
		s.True(updatedWakeUpTime1.Equal(queryResult))
	}, time.Minute*11)

	// Update wake-up time again
	updatedWakeUpTime2 := env.Now().Add(40 * time.Minute)
	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow(SignalType, updatedWakeUpTime2)
	}, time.Minute*12)

	wakeUpTime := env.Now().Add(30 * time.Minute)
	env.ExecuteWorkflow(Workflow, wakeUpTime)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
	elapsed := env.Now().Sub(start)
	// Check that the wake-up time was updated
	s.Equal(elapsed, 40*time.Minute)
}

func (s *UnitTestSuite) Test_CancelSleep() {
	env := s.NewTestWorkflowEnvironment()
	start := env.Now()
	wakeUpTime := env.Now().Add(30 * time.Minute)
	env.RegisterDelayedCallback(func() {
		env.CancelWorkflow()
	}, time.Minute*10)
	env.ExecuteWorkflow(Workflow, wakeUpTime)

	s.True(env.IsWorkflowCompleted())
	workflowErr := env.GetWorkflowError()
	s.Error(workflowErr)
	s.EqualError(errors.Unwrap(workflowErr), "canceled")
	elapsed := env.Now().Sub(start)
	s.Equal(elapsed, 10*time.Minute)
}
