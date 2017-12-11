package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/cadence/testsuite"
	"go.uber.org/cadence/workflow"
)

type UnitTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
}

func TestUnitTestSuite(t *testing.T) {
	suite.Run(t, new(UnitTestSuite))
}

func (s *UnitTestSuite) Test_CronWorkflow_SmallCount() {
	env := s.NewTestWorkflowEnvironment()
	env.OnActivity(sampleCronActivity, mock.Anything, mock.Anything).Return(nil).Times(3)
	env.ExecuteWorkflow(SampleCronWorkflow, ScheduleSpec{JobCount: 3, ScheduleInterval: time.Hour})

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
	env.AssertExpectations(s.T())
}

func (s *UnitTestSuite) Test_CronWorkflow_LargeCount() {
	env := s.NewTestWorkflowEnvironment()
	env.OnActivity(sampleCronActivity, mock.Anything, mock.Anything).Return(nil).Times(10)
	env.ExecuteWorkflow(SampleCronWorkflow, ScheduleSpec{JobCount: 20, ScheduleInterval: time.Hour})

	s.True(env.IsWorkflowCompleted())
	s.NotNil(env.GetWorkflowError())
	_, ok := env.GetWorkflowError().(*workflow.ContinueAsNewError)
	s.True(ok)
	env.AssertExpectations(s.T())
}
