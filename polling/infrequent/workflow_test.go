package infrequent

import (
	"github.com/stretchr/testify/suite"
	"github.com/temporalio/samples-go/polling"
	"go.temporal.io/sdk/testsuite"
	"testing"
)

type UnitTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
}

func TestUnitTestSuite(t *testing.T) {
	suite.Run(t, new(UnitTestSuite))
}

func (s *UnitTestSuite) Test_InfrequentPollingWorkflow() {
	env := s.NewTestWorkflowEnvironment()
	testService := polling.NewTestService(1) // Horrible workaround to avoid a timeout
	a := &PollingActivities{
		TestService: &testService,
	}
	env.RegisterActivity(a)

	env.ExecuteWorkflow(InfrequentPolling)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
	var pollResult string
	s.NoError(env.GetWorkflowResult(&pollResult))
	s.Equalf(pollResult, "OK", "The polling has returned the wrong result")
	env.AssertExpectations(s.T())
}
