package schedule

import (
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
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

func (s *UnitTestSuite) Test_ScheduleWorkflow() {
	env := s.NewTestWorkflowEnvironment()

	env.RegisterWorkflow(SampleScheduleWorkflow)
	env.RegisterActivity(DoSomething)

	err := env.SetSearchAttributesOnStart(map[string]interface{}{
		"TemporalScheduledById":      "schedule_test_ID",
		"TemporalScheduledStartTime": time.Now(),
	})
	s.NoError(err)

	env.OnActivity(DoSomething, mock.Anything, mock.Anything, mock.Anything).Return(nil).Times(3)

	env.ExecuteWorkflow(SampleScheduleWorkflow)

	s.True(env.IsWorkflowCompleted())
	err = env.GetWorkflowError()
	s.NoError(err)
}
