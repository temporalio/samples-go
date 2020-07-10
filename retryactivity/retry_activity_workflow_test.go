package retryactivity

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/activity"
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
	env.RegisterActivity(BatchProcessingActivity)
	var startedIDs []int
	env.OnActivity(BatchProcessingActivity, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(func(ctx context.Context, firstTaskID, batchSize int, processDelay time.Duration) error {
			i := firstTaskID
			if activity.HasHeartbeatDetails(ctx) {
				var completedIdx int
				if err := activity.GetHeartbeatDetails(ctx, &completedIdx); err == nil {
					i = completedIdx + 1
				}
			}
			startedIDs = append(startedIDs, i)

			return BatchProcessingActivity(ctx, firstTaskID, batchSize, time.Nanosecond /* override for test */)
		})
	env.ExecuteWorkflow(RetryWorkflow)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
	s.Equal([]int{0, 6, 12, 18}, startedIDs)
	env.AssertExpectations(s.T())
}
