package timer

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
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

func (s *UnitTestSuite) Test_Workflow_FastProcessing() {
	env := s.NewTestWorkflowEnvironment()

	// mock to return immediately to simulate fast processing case
	env.OnActivity(OrderProcessingActivity, mock.Anything).Return(nil)
	env.OnActivity(SendEmailActivity, mock.Anything).Return(func(ctx context.Context) error {
		// in fast processing case, this method should not get called
		s.FailNow("SendEmailActivity should not get called")
		return nil
	})

	env.ExecuteWorkflow(SampleTimerWorkflow, time.Minute)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
}

func (s *UnitTestSuite) Test_Workflow_SlowProcessing() {
	env := s.NewTestWorkflowEnvironment()

	wg := &sync.WaitGroup{}
	wg.Add(1)
	env.OnActivity(OrderProcessingActivity, mock.Anything).Return(func(ctx context.Context) error {
		// simulate slow processing, will complete this activity only after the SendEmailActivity is called.
		wg.Wait()
		return nil
	})
	env.OnActivity(SendEmailActivity, mock.Anything).Return(func(ctx context.Context) error {
		wg.Done()
		return nil
	})

	env.ExecuteWorkflow(SampleTimerWorkflow, time.Microsecond)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
}
