package periodic_sequence

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/temporalio/samples-go/polling"

	"go.temporal.io/sdk/testsuite"
)

func Test_PeriodicPollingWorkflow(t *testing.T) {
	s := testsuite.WorkflowTestSuite{}
	env := s.NewTestWorkflowEnvironment()
	testService := polling.NewTestService(5)
	a := &PollingActivities{
		TestService: &testService,
	}
	env.RegisterActivity(a)
	env.RegisterWorkflow(PollingChildWorkflow)

	env.ExecuteWorkflow(PeriodicSequencePolling, 100*time.Millisecond)

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())

	env.AssertExpectations(t)
}
