package frequent

import (
	"time"

	"github.com/temporalio/samples-go/polling"

	"go.temporal.io/sdk/testsuite"
)

func TestFrequentPollingWorkflow() {
	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestWorkflowEnvironment()
	testService := polling.NewTestService(5)
	a := &PollingActivities{
		TestService:  &testService,
		PollInterval: 100 * time.Millisecond, // Beware of test timeouts if you change this
	}
	env.RegisterActivity(a)
	env.ExecuteWorkflow(FrequentPolling)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
	var pollResult string
	s.NoError(env.GetWorkflowResult(&pollResult))
	s.Equalf(pollResult, "OK", "The polling has returned the wrong result")

	env.AssertExpectations(s.T())
}
