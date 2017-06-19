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

func (s *UnitTestSuite) Test_ExclusiveChoiceWorkflow() {
	env := s.NewTestWorkflowEnvironment()
	env.ExecuteWorkflow(ExclusiveChoiceWorkflow)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
}

func (s *UnitTestSuite) Test_MultiChoiceWorkflow() {
	env := s.NewTestWorkflowEnvironment()
	env.ExecuteWorkflow(MultiChoiceWorkflow)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
}
