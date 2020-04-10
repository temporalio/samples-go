package fileprocessing

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"go.temporal.io/temporal/testsuite"
)

type UnitTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
}

func TestUnitTestSuite(t *testing.T) {
	suite.Run(t, new(UnitTestSuite))
}

func (s *UnitTestSuite) Test_SampleFileProcessingWorkflow() {
	fileID := "test-file-id"
	expectedCall := []string{
		"downloadFileActivity",
		"processFileActivity",
		"uploadFileActivity",
	}

	var activityCalled []string
	env := s.NewTestWorkflowEnvironment()

	env.RegisterActivity(&Activities{&BlobStore{}})

	env.ExecuteWorkflow(SampleFileProcessingWorkflow, fileID)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
	s.Equal(expectedCall, activityCalled)
}
