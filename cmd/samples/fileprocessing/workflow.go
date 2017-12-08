package main

import (
	"time"

	"github.com/pborman/uuid"

	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

type (
	fileInfo struct {
		FileName string
		HostID   string
	}
)

// ApplicationName is the task list for this sample
const ApplicationName = "FileProcessorGroup"

// HostID - Use a new uuid just for demo so we can run 2 host specific activity workers on same machine.
// In real world case, you would use a hostname or ip address as HostID.
var HostID = ApplicationName + "_" + uuid.New()

// This is registration process where you register all your workflow handlers.
func init() {
	workflow.Register(SampleFileProcessingWorkflow)
}

//SampleFileProcessingWorkflow workflow decider
func SampleFileProcessingWorkflow(ctx workflow.Context, fileID string) (err error) {
	// step 1: download resource file
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
		HeartbeatTimeout:       time.Second * 20,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var fInfo *fileInfo
	err = workflow.ExecuteActivity(ctx, downloadFileActivity, fileID).Get(ctx, &fInfo)
	if err != nil {
		workflow.GetLogger(ctx).Error("Workflow failed.", zap.String("Error", err.Error()))
		return err
	}

	// following activities needs to be run on the same host as first activity, through this host specific tasklist.
	// HostSpecificGroupList and with a shorter queue timeout.
	hCtx := workflow.WithTaskList(ctx, fInfo.HostID)
	hCtx = workflow.WithScheduleToStartTimeout(ctx, time.Second*10)

	// step 2: process file. We use simple retry strategy to retry on queue timeout error
	var fInfoProcessed *fileInfo
	err = retryOnQueueTimeout(hCtx, &fInfoProcessed, processFileActivity, *fInfo)
	if err != nil {
		workflow.GetLogger(ctx).Error("Workflow failed.", zap.String("Error", err.Error()))
		return err
	}

	// step 3: upload processed file.
	err = retryOnQueueTimeout(hCtx, nil, uploadFileActivity, *fInfoProcessed)
	if err != nil {
		workflow.GetLogger(ctx).Error("Workflow failed.", zap.String("Error", err.Error()))
		return err
	}

	workflow.GetLogger(ctx).Info("Workflow completed.")
	return nil
}

func retryOnQueueTimeout(ctx workflow.Context, result interface{}, fn interface{}, args ...interface{}) (err error) {
	for i := 1; i < 5; i++ {
		future := workflow.ExecuteActivity(ctx, fn, args...)
		// wait until it is done, but we don't care about the result yet.
		err = future.Get(ctx, result)
		if err != nil {
			// try again
			continue
		}
		return nil
	}

	// we are not able to make it with all retries, so give up
	return err
}
