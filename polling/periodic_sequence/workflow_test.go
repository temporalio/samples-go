package periodic_sequence

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

func (s *UnitTestSuite) Test_PeriodicPollingWorkflow() {
	env := s.NewTestWorkflowEnvironment()
	testService := polling.NewTestService(5)
	a := &PollingActivities{
		TestService: &testService,
	}
	env.RegisterActivity(a)
	env.RegisterWorkflow(PollingChildWorkflow)

	env.ExecuteWorkflow(PeriodicSequencePolling, 100*time.Millisecond)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())

	env.AssertExpectations(s.T())
}
