package main

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"go.uber.org/cadence"
)

type UnitTestSuite struct {
	suite.Suite
	cadence.WorkflowTestSuite
}

func TestUnitTestSuite(t *testing.T) {
	suite.Run(t, new(UnitTestSuite))
}

func (s *UnitTestSuite) Test_Workflow() {
	env := s.NewTestWorkflowEnvironment()
	env.OverrideActivity(sampleActivity, func(ctx context.Context, currentBranchID int, totalDuration time.Duration) (string, error) {
		// make branch 0 super fast so we don't have to wait sleep time in unit test
		if currentBranchID == 0 {
			totalDuration = time.Nanosecond
		}
		return sampleActivity(ctx, currentBranchID, totalDuration)
	})
	env.ExecuteWorkflow(SamplePickFirstWorkflow)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
}
