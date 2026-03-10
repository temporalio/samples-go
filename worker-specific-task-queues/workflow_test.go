package worker_specific_task_queues

import (
	"context"
	"errors"
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

func Test_RetrySuccess(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	// Mock activity implementation
	var a WorkerSpecificTaskQueue
	env.RegisterActivityWithOptions(a.GetWorkerSpecificTaskQueue, activity.RegisterOptions{
		Name: "GetWorkerSpecificTaskQueue",
	})

	counter := 0
	env.OnActivity("GetWorkerSpecificTaskQueue", mock.Anything).Return(func(ctx context.Context) (string, error) {
		counter++
		// Workflow retries up to 5 times
		if counter < 3 {
			return "", errors.New("temporary error")
		}
		return "unique-task-queue", nil
	})
	env.OnActivity(DownloadFile, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	env.OnActivity(ProcessFile, mock.Anything, mock.Anything).Return(nil)
	env.OnActivity(DeleteFile, mock.Anything, mock.Anything).Return(nil)

	env.ExecuteWorkflow(FileProcessingWorkflow)

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
}

func Test_RetryFail(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	// Mock activity implementation
	var a WorkerSpecificTaskQueue
	env.RegisterActivityWithOptions(a.GetWorkerSpecificTaskQueue, activity.RegisterOptions{
		Name: "GetWorkerSpecificTaskQueue",
	})
	env.OnActivity("GetWorkerSpecificTaskQueue", mock.Anything).Return(func(ctx context.Context) (string, error) {
		return "", errors.New("error to show a retry mechanic failure")
	})
	env.OnActivity(DownloadFile, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	env.OnActivity(ProcessFile, mock.Anything, mock.Anything).Return(nil)
	env.OnActivity(DeleteFile, mock.Anything, mock.Anything).Return(nil)

	env.ExecuteWorkflow(FileProcessingWorkflow)

	require.True(t, env.IsWorkflowCompleted())
	require.Error(t, env.GetWorkflowError())
}
