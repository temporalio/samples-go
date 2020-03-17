package main

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

func (s *UnitTestSuite) Test_ExclusiveChoiceWorkflow() {
	env := s.NewTestWorkflowEnvironment()
	env.RegisterActivity(getOrderActivity)
	env.RegisterActivity(orderAppleActivity)
	env.RegisterActivity(orderBananaActivity)
	env.RegisterActivity(orderCherryActivity)
	env.RegisterActivity(orderOrangeActivity)

	env.ExecuteWorkflow(ExclusiveChoiceWorkflow)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
}

func (s *UnitTestSuite) Test_MultiChoiceWorkflow() {
	env := s.NewTestWorkflowEnvironment()
	env.RegisterActivity(getBasketOrderActivity)

	env.ExecuteWorkflow(MultiChoiceWorkflow)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
}
