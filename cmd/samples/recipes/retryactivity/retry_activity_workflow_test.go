package main

import (
	"context"
	"testing"

	"github.com/pkg/errors"
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
	maxRetry := 5
	retryCount := 0
	env.OverrideActivity(sampleActivity, func(ctx context.Context) error {
		retryCount++
		if retryCount < maxRetry {
			return errors.New("failed, please retry")
		}
		return nil
	})
	env.ExecuteWorkflow(RetryWorkflow, maxRetry)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
	s.Equal(maxRetry, retryCount)
}
