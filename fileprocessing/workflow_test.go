package fileprocessing

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

func (s *UnitTestSuite) Test_SampleFileProcessingWorkflow() {
	env := s.NewTestWorkflowEnvironment()
	env.SetWorkerOptions(worker.Options{
		EnableSessionWorker: true, // Important for a worker to participate in the session
	})
	var a *Activities

	env.OnActivity(a.DownloadFileActivity, mock.Anything, "file1").Return("file2", nil)
	env.OnActivity(a.ProcessFileActivity, mock.Anything, "file2").Return("file3", nil)
	env.OnActivity(a.UploadFileActivity, mock.Anything, "file3").Return(nil)

	env.RegisterActivity(a)

	env.ExecuteWorkflow(SampleFileProcessingWorkflow, "file1")

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())

	env.AssertExpectations(s.T())
}
