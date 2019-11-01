package main

import (
	"time"

	"github.com/pborman/uuid"
	"go.temporal.io/temporal"
	"go.temporal.io/temporal/workflow"
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
		ScheduleToStartTimeout: time.Second * 5,
		StartToCloseTimeout:    time.Minute,
		HeartbeatTimeout:       time.Second * 2, // such a short timeout to make sample fail over very fast
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:          time.Second,
			BackoffCoefficient:       2.0,
			MaximumInterval:          time.Minute,
			ExpirationInterval:       time.Minute * 10,
			NonRetriableErrorReasons: []string{"bad-error"},
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	// Retry the whole sequence from the first activity on any error
	// to retry it on a different host. In a real application it might be reasonable to
	// retry individual activities and the whole sequence discriminating between different types of errors.
	// See the retryactivity sample for a more sophisticated retry implementation.
	for i := 1; i < 5; i++ {
		err = processFile(ctx, fileID)
		if err == nil {
			break
		}
	}
	if err != nil {
		workflow.GetLogger(ctx).Error("Workflow failed.", zap.String("Error", err.Error()))
	} else {
		workflow.GetLogger(ctx).Info("Workflow completed.")
	}
	return err
}

func processFile(ctx workflow.Context, fileID string) (err error) {
	var fInfo *fileInfo
	so := &workflow.SessionOptions{
		CreationTimeout:  time.Minute,
		ExecutionTimeout: time.Minute,
	}

	sessionCtx, err := workflow.CreateSession(ctx, so)
	if err != nil {
		return err
	}
	defer workflow.CompleteSession(sessionCtx)

	err = workflow.ExecuteActivity(sessionCtx, downloadFileActivityName, fileID).Get(sessionCtx, &fInfo)
	if err != nil {
		return err
	}

	var fInfoProcessed *fileInfo
	err = workflow.ExecuteActivity(sessionCtx, processFileActivityName, *fInfo).Get(sessionCtx, &fInfoProcessed)
	if err != nil {
		return err
	}

	err = workflow.ExecuteActivity(sessionCtx, uploadFileActivityName, *fInfoProcessed).Get(sessionCtx, nil)
	return err
}
