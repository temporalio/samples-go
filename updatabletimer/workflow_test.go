package updatabletimer

import (
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

func (s *UnitTestSuite) Test_Workflow() {
	env := s.NewTestWorkflowEnvironment()
	start := env.Now()
	wakeUpTime := env.Now().Add(30 * time.Minute)
	env.RegisterDelayedCallback(func() {
		value, err := env.QueryWorkflow(QueryType)
		s.NoError(err)
		var queryResult time.Time
		err = value.Get(&queryResult)
		s.NoError(err)
		s.Equal(wakeUpTime, queryResult)
	}, time.Minute*10)
	env.ExecuteWorkflow(Workflow, wakeUpTime)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
	elapsed := env.Now().Sub(start)
	s.True(elapsed >= 30 * time.Minute)
	s.True(elapsed < 31 * time.Minute)
}

func (s *UnitTestSuite) Test_Workflow_UpdateWakeUpTime() {
	env := s.NewTestWorkflowEnvironment()
	start := env.Now()
	updatedWakeUpTime := env.Now().Add(15 * time.Minute)
	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow(SignalType, updatedWakeUpTime)
	}, time.Minute*10)
	wakeUpTime := env.Now().Add(30 * time.Minute)
	env.ExecuteWorkflow(Workflow, wakeUpTime)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
	elapsed := env.Now().Sub(start).Milliseconds()
	s.True(elapsed >= 15*60000)
	s.True(elapsed < 16*60000)
}

//func (s *UnitTestSuite) Test_Workflow_SlowProcessing() {
//	env := s.NewTestWorkflowEnvironment()
//
//	wg := &sync.WaitGroup{}
//	wg.Add(1)
//	env.OnActivity(OrderProcessingActivity, mock.Anything).Return(func(ctx context.Context) error {
//		// simulate slow processing, will complete this activity only after the SendEmailActivity is called.
//		wg.Wait()
//		return nil
//	})
//	env.OnActivity(SendEmailActivity, mock.Anything).Return(func(ctx context.Context) error {
//		wg.Done()
//		return nil
//	})
//
//	env.ExecuteWorkflow(Workflow, time.Microsecond)
//
//	s.True(env.IsWorkflowCompleted())
//	s.NoError(env.GetWorkflowError())
//}
