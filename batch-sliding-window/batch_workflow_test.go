package batch_sliding_window

import (
	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/testsuite"
	"testing"
)

type UnitTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
}

func TestUnitTestSuite(t *testing.T) {
	suite.Run(t, new(UnitTestSuite))
}

func (s *UnitTestSuite) Test_ProcessBatchWorkflow() {
	env := s.NewTestWorkflowEnvironment()
	env.RegisterActivity(&RecordLoader{RecordCount: 1})
	env.ExecuteWorkflow(ProcessBatchWorkflow, &ProcessBatchWorkflowInput{
		PageSize:          1,
		SlidingWindowSize: 1,
		Partitions:        1,
	})
	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
	var result int
	s.NoError(env.GetWorkflowResult(&result))
	s.Equal(1, result)
}
