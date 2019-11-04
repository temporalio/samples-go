package main

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
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

func (s *UnitTestSuite) Test_Workflow_FastProcessing() {
	env := s.NewTestWorkflowEnvironment()

	// mock to return immediately to simulate fast processing case
	env.OnActivity(orderProcessingActivity, mock.Anything).Return(nil)
	env.OnActivity(sendEmailActivity, mock.Anything).Return(func(ctx context.Context) error {
		// in fast processing case, this method should not get called
		s.FailNow("sendEmailActivity should not get called")
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
	env.OnActivity(orderProcessingActivity, mock.Anything).Return(func(ctx context.Context) error {
		// simulate slow processing, will complete this activity only after the sendEmailActivity is called.
		wg.Wait()
		return nil
	})
	env.OnActivity(sendEmailActivity, mock.Anything).Return(func(ctx context.Context) error {
		wg.Done()
		return nil
	})

	env.ExecuteWorkflow(SampleTimerWorkflow, time.Microsecond)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
}
