package choice

import (
	"fmt"
	"testing"

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

func (s *UnitTestSuite) Test_ExclusiveChoiceWorkflowSucceeds() {
	env := s.NewTestWorkflowEnvironment()

	orderChoices := []string{
		OrderChoiceApple,
		OrderChoiceOrange}
	env.RegisterActivity(&OrderActivities{OrderChoices: orderChoices})

	env.ExecuteWorkflow(ExclusiveChoiceWorkflow)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
}

func (s *UnitTestSuite) Test_ExclusiveChoiceWorkflowFailOnGetOrderFailure() {
	env := s.NewTestWorkflowEnvironment()
	activities := &OrderActivities{}
	env.OnActivity(activities.GetOrder, mock.Anything).Return("", fmt.Errorf("Get Order Error"))

	env.ExecuteWorkflow(ExclusiveChoiceWorkflow)

	s.True(env.IsWorkflowCompleted())
	s.Error(env.GetWorkflowError())
}

func (s *UnitTestSuite) Test_ExclusiveChoiceWorkflowFailOnOrdering() {
	env := s.NewTestWorkflowEnvironment()
	orderChoices := []string{
		OrderChoiceApple,
	}
	activities := &OrderActivities{orderChoices}
	env.RegisterActivity(activities.GetOrder)
	env.OnActivity(activities.OrderOrange, mock.Anything).Return(nil, fmt.Errorf("Get Order Error"))

	env.ExecuteWorkflow(ExclusiveChoiceWorkflow)

	s.True(env.IsWorkflowCompleted())
	s.Error(env.GetWorkflowError())
}
