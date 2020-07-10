package parallel

import (
	"sort"
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

func (s *UnitTestSuite) Test_ParallelWorkflow() {
	env := s.NewTestWorkflowEnvironment()
	env.RegisterActivity(SampleActivity)
	env.OnActivity(SampleActivity, mock.Anything).Return("one", nil).Once()
	env.OnActivity(SampleActivity, mock.Anything).Return("two", nil).Once()
	env.OnActivity(SampleActivity, mock.Anything).Return("three", nil).Once()
	env.ExecuteWorkflow(SampleParallelWorkflow)
	var result []string
	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
	s.NoError(env.GetWorkflowResult(&result))

	expected := []string{"one", "three", "two"}
	sort.Strings(expected)
	sort.Strings(result)
	s.Equal(expected, result)
}
