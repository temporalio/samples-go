package largeeventhistory

import (
	"testing"

	"github.com/stretchr/testify/mock"
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

func (s *UnitTestSuite) Test_LargePayloadWorkflow() {
	env := s.NewTestWorkflowEnvironment()
	env.SetWorkerOptions(worker.Options{})
	var a *Activities

	data := []byte{}
	env.OnActivity(a.CreateLargeResultActivity, mock.Anything, 1*1024).Return(data, nil)
	env.OnActivity(a.ProcessLargeInputActivity, mock.Anything, data).Return(nil)

	env.RegisterActivity(a)

	env.ExecuteWorkflow(LargePayloadWorkflow, 1*1024)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
}
