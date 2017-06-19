package main

import (
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

func (s *UnitTestSuite) Test_BranchWorkflow() {
	env := s.NewTestWorkflowEnvironment()
	env.ExecuteWorkflow(SampleBranchWorkflow)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
}

func (s *UnitTestSuite) Test_ParallelWorkflow() {
	env := s.NewTestWorkflowEnvironment()
	env.ExecuteWorkflow(SampleParallelWorkflow)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
}
