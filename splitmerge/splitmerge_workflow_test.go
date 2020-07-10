package splitmerge

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

func (s *UnitTestSuite) Test_Workflow() {
	env := s.NewTestWorkflowEnvironment()
	env.RegisterActivity(ChunkProcessingActivity)

	workerCount := 5
	env.ExecuteWorkflow(SampleSplitMergeWorkflow, workerCount)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())

	var result ChunkResult
	_ = env.GetWorkflowResult(&result)

	totalItem, totalSum := 0, 0
	for i := 1; i <= workerCount; i++ {
		totalItem += i
		totalSum += i * i
	}

	s.Equal(totalItem, result.NumberOfItemsInChunk)
	s.Equal(totalSum, result.SumInChunk)
}
