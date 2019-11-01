package main

import (
	"context"
	"go.temporal.io/temporal/activity"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
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

func (s *UnitTestSuite) Test_Workflow() {
	env := s.NewTestWorkflowEnvironment()
	var startedIDs []int
	env.OnActivity(batchProcessingActivity, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(func(ctx context.Context, firstTaskID, batchSize int, processDelay time.Duration) error {
			i := firstTaskID
			if activity.HasHeartbeatDetails(ctx) {
				var completedIdx int
				if err := activity.GetHeartbeatDetails(ctx, &completedIdx); err == nil {
					i = completedIdx + 1
				}
			}
			startedIDs = append(startedIDs, i)

			return batchProcessingActivity(ctx, firstTaskID, batchSize, time.Nanosecond /* override for test */)
		})
	env.ExecuteWorkflow(RetryWorkflow)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
	s.Equal([]int{0, 6, 12, 18}, startedIDs)
	env.AssertExpectations(s.T())
}
