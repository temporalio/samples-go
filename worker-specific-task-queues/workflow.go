package worker_specific_task_queues

import (
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"go.temporal.io/sdk/workflow"
)

// FileProcessingWorkflow is a workflow that uses Worker-specific Task Queues to run multiple Activities on a consistent
// host.
func FileProcessingWorkflow(ctx workflow.Context) (err error) {
	for attempt := 1; attempt <= 5; attempt++ {
		if err = processFile(ctx); err == nil {
			workflow.GetLogger(ctx).Info("Workflow completed.")
			return
		}
		workflow.GetLogger(ctx).Error("Attempt failed, trying on new worker", attempt)
	}
	workflow.GetLogger(ctx).Error("Workflow failed after multiple retries.", "Error", err.Error())
	return
}

func processFile(ctx workflow.Context) (err error) {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
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
