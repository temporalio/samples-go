package ctxpropagation

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	commonpb "go.temporal.io/api/common/v1"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
)

type UnitTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
}

func TestUnitTestSuite(t *testing.T) {
	s := &UnitTestSuite{}
	// Create header as if it was injected from context.
	// Test suite doesn't accept context therefore it is not possible to inject PropagateKey value it from real context.
	payload, _ := converter.GetDefaultDataConverter().ToPayload(Values{"some key", "some value"})
	s.SetHeader(&commonpb.Header{
		Fields: map[string]*commonpb.Payload{
			propagationKey: payload,
		},
	})

	suite.Run(t, s)
}

func (s *UnitTestSuite) Test_CtxPropWorkflow() {
	env := s.NewTestWorkflowEnvironment()
	env.SetContextPropagators([]workflow.ContextPropagator{NewContextPropagator()})
	env.RegisterActivity(SampleActivity)

	var propagatedValue interface{}
	env.SetOnActivityStartedListener(func(activityInfo *activity.Info, ctx context.Context, args converter.EncodedValues) {
		// PropagateKey should be propagated by custom context propagator from propagationKey header.
		propagatedValue = ctx.Value(PropagateKey)
	})

	env.ExecuteWorkflow(CtxPropWorkflow)
	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())

	s.NotNil(propagatedValue)
	pv, ok := propagatedValue.(Values)
	s.True(ok)
	s.Equal("some key", pv.Key)
	s.Equal("some value", pv.Value)
}
