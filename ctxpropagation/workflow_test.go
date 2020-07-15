package ctxpropagation

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/encoded"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
)

type UnitTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
}

func TestUnitTestSuite(t *testing.T) {
	suite.Run(t, new(UnitTestSuite))
}

// TODO: Modify this unit test to actually test that propagation is happening.
// This will be possible after https://github.com/temporalio/go-sdk/issues/190 is resolved.
func (s *UnitTestSuite) Test_CtxPropWorkflow() {
	s.SetContextPropagators([]workflow.ContextPropagator{NewContextPropagator()})
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
