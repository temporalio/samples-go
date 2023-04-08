package frequent

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/temporalio/samples-go/polling"

	"go.temporal.io/sdk/testsuite"
)

func TestFrequentPollingWorkflow(t *testing.T) {
	var s testsuite.WorkflowTestSuite
	env := s.NewTestWorkflowEnvironment()
	testService := polling.NewTestService(5)
	a := &PollingActivities{
		TestService:  &testService,
		PollInterval: 100 * time.Millisecond, // Beware of test timeouts if you change this
	}
	env.RegisterActivity(a)
	env.ExecuteWorkflow(FrequentPolling)

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	var pollResult string
	require.NoError(t, env.GetWorkflowResult(&pollResult))
	require.Equalf(t, pollResult, "OK", "The polling has returned the wrong result")

	env.AssertExpectations(t)
}
