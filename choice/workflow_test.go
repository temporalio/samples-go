package choice

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
	env.RegisterActivity(GetOrderActivity)
	env.RegisterActivity(OrderAppleActivity)
	env.RegisterActivity(OrderBananaActivity)
	env.RegisterActivity(OrderCherryActivity)
	env.RegisterActivity(OrderOrangeActivity)

	env.ExecuteWorkflow(ExclusiveChoiceWorkflow)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
}

func (s *UnitTestSuite) Test_MultiChoiceWorkflow() {
	env := s.NewTestWorkflowEnvironment()
	env.RegisterActivity(GetBasketOrderActivity)

	env.ExecuteWorkflow(MultiChoiceWorkflow)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
}
