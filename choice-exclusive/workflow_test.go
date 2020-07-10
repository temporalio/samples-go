package choice

import (
	"testing"

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

func (s *UnitTestSuite) Test_ExclusiveChoiceWorkflow() {
	env := s.NewTestWorkflowEnvironment()

	orderChoices := []string{
		OrderChoiceApple,
		OrderChoiceOrange}
	env.RegisterActivity(&OrderActivities{OrderChoices: orderChoices})

	env.ExecuteWorkflow(ExclusiveChoiceWorkflow)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
}
