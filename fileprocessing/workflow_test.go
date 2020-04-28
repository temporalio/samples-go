package fileprocessing

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.temporal.io/temporal/activity"
	"go.temporal.io/temporal/encoded"
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

	env.RegisterActivityWithOptions(
		DownloadFileActivity,
		activity.RegisterOptions{Name: DownloadFileActivityName},
	)
	env.RegisterActivityWithOptions(
		ProcessFileActivity,
		activity.RegisterOptions{Name: ProcessFileActivityName},
	)
	env.RegisterActivityWithOptions(
		UploadFileActivity,
		activity.RegisterOptions{Name: UploadFileActivityName},
	)

	env.SetOnActivityStartedListener(func(activityInfo *activity.Info, ctx context.Context, args encoded.Values) {
		activityType := activityInfo.ActivityType.Name
		if strings.HasPrefix(activityType, "internalSession") {
			return
		}
		activityCalled = append(activityCalled, activityType)
		switch activityType {
		case expectedCall[0]:
			var input string
			s.NoError(args.Get(&input))
			s.Equal(fileID, input)
		case expectedCall[1]:
			var input string
			s.NoError(args.Get(&input))
		case expectedCall[2]:
			var input string
			s.NoError(args.Get(&input))
		default:
			panic("unexpected activity call")
		}
	})
	env.ExecuteWorkflow(SampleFileProcessingWorkflow, fileID)

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
	s.Equal(expectedCall, activityCalled)
}
