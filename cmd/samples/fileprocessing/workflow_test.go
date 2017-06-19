package main

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/cadence"
)

type UnitTestSuite struct {
	suite.Suite
	cadence.WorkflowTestSuite
}

func TestUnitTestSuite(t *testing.T) {
	suite.Run(t, new(UnitTestSuite))
}

func (s *UnitTestSuite) Test_SampleFileProcessingWorkflow() {
	fileID := "test-file-id"
	expectedCall := []string{
		"github.com/samarabbas/cadence-samples/cmd/samples/fileprocessing.downloadFileActivity",
		"github.com/samarabbas/cadence-samples/cmd/samples/fileprocessing.processFileActivity",
		"github.com/samarabbas/cadence-samples/cmd/samples/fileprocessing.uploadFileActivity",
	}

	var activityCalled []string
	env := s.NewTestWorkflowEnvironment()
	env.SetOnActivityStartedListener(func(activityInfo *cadence.ActivityInfo, ctx context.Context, args cadence.EncodedValues) {
		activityType := activityInfo.ActivityType.Name
		activityCalled = append(activityCalled, activityType)
		switch activityType {
		case expectedCall[0]:
			var input string
			s.NoError(args.Get(&input))
			s.Equal(fileID, input)
		case expectedCall[1]:
			var input fileInfo
			s.NoError(args.Get(&input))
			s.Equal(input.HostID, HostID)
		case expectedCall[2]:
			var input fileInfo
			s.NoError(args.Get(&input))
			s.Equal(input.HostID, HostID)
		default:
			panic("unexpected activity call: "+activityType)
		}
	})
	env.ExecuteWorkflow(SampleFileProcessingWorkflow, fileID)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
	s.Equal(expectedCall, activityCalled)
}
