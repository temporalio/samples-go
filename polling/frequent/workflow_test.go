package frequent

import (
	"github.com/temporalio/samples-go/polling"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/testsuite"
)

type UnitTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
}

func TestUnitTestSuite(t *testing.T) {
	suite.Run(t, new(UnitTestSuite))
}

func (s *UnitTestSuite) Test_FrequentPollingWorkflow() {
	env := s.NewTestWorkflowEnvironment()
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
