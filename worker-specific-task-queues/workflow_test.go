package worker_specific_task_queues

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
	var a WorkerSpecificTaskQueue
	env.RegisterActivityWithOptions(a.GetWorkerSpecificTaskQueue, activity.RegisterOptions{
		Name: "GetWorkerSpecificTaskQueue",
	})
	env.OnActivity("GetWorkerSpecificTaskQueue", mock.Anything).Return("unique-task-queue", nil)
	env.OnActivity(DownloadFile, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	env.OnActivity(ProcessFile, mock.Anything, mock.Anything).Return(nil)
	env.OnActivity(DeleteFile, mock.Anything, mock.Anything).Return(nil)

	env.ExecuteWorkflow(FileProcessingWorkflow)

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
}
