package worker_specific_task_queues

import (
	"go.temporal.io/sdk/temporal"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"go.temporal.io/sdk/workflow"
)

// FileProcessingWorkflow is a workflow that uses Worker-specific Task Queues to run multiple Activities on a consistent
// host.
func FileProcessingWorkflow(ctx workflow.Context) (err error) {
	// When using a worker-specific task queue, if a failure occurs, we want to retry all of the worker-specific
	// logic, so wrap all the logic here in a loop.
	for attempt := range 5 {
		if err = processFile(ctx); err == nil {
			workflow.GetLogger(ctx).Info("Workflow completed.")
			return
		}
		workflow.GetLogger(ctx).Error("Attempt failed, trying on new worker", attempt+1)
	}
	workflow.GetLogger(ctx).Error("Workflow failed after multiple retries.", "Error", err.Error())
	return
}

func processFile(ctx workflow.Context) (err error) {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 1,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	var WorkerSpecificTaskQueue string
	err = workflow.ExecuteActivity(ctx, "GetWorkerSpecificTaskQueue").Get(ctx, &WorkerSpecificTaskQueue)
	if err != nil {
		return
	}
	ao = workflow.ActivityOptions{
		// Note the use of scheduleToCloseTimeout.
		// The reason this timeout type is used is because this task queue is unique
		// to a single worker. When that worker goes away, there won't be a way for these
		// activities to progress.
		ScheduleToCloseTimeout: time.Minute,

		TaskQueue: WorkerSpecificTaskQueue,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	downloadPath := filepath.Join("/tmp", uuid.New().String())
	err = workflow.ExecuteActivity(ctx, DownloadFile, "https://temporal.io", downloadPath).Get(ctx, nil)
	if err != nil {
		return
	}
	defer func() {
		err = workflow.ExecuteActivity(ctx, DeleteFile, downloadPath).Get(ctx, nil)
	}()

	err = workflow.ExecuteActivity(ctx, ProcessFile, downloadPath).Get(ctx, nil)
	return
}
