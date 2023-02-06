package activities_sticky_queues

import (
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/testsuite"
)

func Test_Workflow(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	// Mock activity implementation
	var a StickyTaskQueue
	env.RegisterActivityWithOptions(a.GetStickyTaskQueue, activity.RegisterOptions{
		Name: "GetStickyTaskQueue",
	})
	env.OnActivity("GetStickyTaskQueue", mock.Anything).Return("unique-sticky-task-queue", nil)
	env.OnActivity(DownloadFile, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	env.OnActivity(ProcessFile, mock.Anything, mock.Anything).Return(nil)
	env.OnActivity(DeleteFile, mock.Anything, mock.Anything).Return(nil)

	env.ExecuteWorkflow(FileProcessingWorkflow)

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
}
