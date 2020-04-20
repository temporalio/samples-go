package ctxpropagation

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.temporal.io/temporal/activity"
	"go.temporal.io/temporal/encoded"
	"go.temporal.io/temporal/testsuite"
)

type UnitTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
}

func TestUnitTestSuite(t *testing.T) {
	suite.Run(t, new(UnitTestSuite))
}

func (s *UnitTestSuite) Test_CtxPropWorkflow() {
	env := s.NewTestWorkflowEnvironment()
	env.RegisterActivity(SampleActivity)

	var activityCalled []string
	env.SetOnActivityStartedListener(func(activityInfo *activity.Info, ctx context.Context, args encoded.Values) {
		activityType := activityInfo.ActivityType.Name
		activityCalled = append(activityCalled, activityType)
	})
	env.ExecuteWorkflow(CtxPropWorkflow)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
	s.Equal([]string{"SampleActivity"}, activityCalled)
}
