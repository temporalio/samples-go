package activities_sticky_queues

import (
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"go.temporal.io/sdk/workflow"
)

// FileProcessingWorkflow is a workflow that uses stick activity queues to process files
// on a consistent host.
func FileProcessingWorkflow(ctx workflow.Context) (err error) {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	var stickyTaskQueue string
	err = workflow.ExecuteActivity(ctx, "GetStickyTaskQueue").Get(ctx, &stickyTaskQueue)
	if err != nil {
		return
	}
	ao = workflow.ActivityOptions{
		// Note the use of scheduleToCloseTimeout.
		// The reason this timeout type is used is because this task queue is unique
		// to a single worker. When that worker goes away, there won't be a way for these
		// activities to progress.
		ScheduleToCloseTimeout: time.Minute,

		TaskQueue: stickyTaskQueue,
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
