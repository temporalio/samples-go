package rainbowstatuses

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/worker"

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

func (s *UnitTestSuite) Test_RainbowStatusesWorkflow() {
	env := s.NewTestWorkflowEnvironment()
	env.SetWorkerOptions(worker.Options{})

	var a *Activities
	env.OnActivity(a.CompletedActivity, mock.Anything).Return(nil)
	env.OnActivity(a.LongActivity, mock.Anything).Return(nil)

	env.RegisterActivity(a)

	status := enums.WORKFLOW_EXECUTION_STATUS_COMPLETED
	env.ExecuteWorkflow(RainbowStatusesWorkflow, status)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
}

func (s *UnitTestSuite) TestReplayWorkflowHistoryFromFile() {
	replayer := worker.NewWorkflowReplayer()

	replayer.RegisterWorkflow(RainbowStatusesWorkflow)

	err := replayer.ReplayWorkflowHistoryFromJSONFile(nil, "rainbowstatusesworkflow.json")
	s.NoError(err)
}
