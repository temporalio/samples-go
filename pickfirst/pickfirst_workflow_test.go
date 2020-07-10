package pickfirst

import (
	"context"
	"testing"
	"time"

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

func (s *UnitTestSuite) Test_Workflow() {
	env := s.NewTestWorkflowEnvironment()
	env.RegisterActivity(SampleActivity)
	env.OnActivity(SampleActivity, mock.Anything, mock.Anything, mock.Anything).
		Return(func(ctx context.Context, currentBranchID int, totalDuration time.Duration) (string, error) {
			// make branch 0 super fast so we don't have to wait sleep time in unit test
			if currentBranchID == 0 {
				totalDuration = time.Nanosecond
			}
			return SampleActivity(ctx, currentBranchID, totalDuration)
		})
	env.ExecuteWorkflow(SamplePickFirstWorkflow)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
	env.AssertExpectations(s.T())
}
