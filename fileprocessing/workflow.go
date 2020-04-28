package fileprocessing

import (
	"time"

	"go.temporal.io/temporal"
	"go.temporal.io/temporal/workflow"
	"go.uber.org/zap"
)

// SampleFileProcessingWorkflow workflow decider
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

func processFile(ctx workflow.Context, fileName string) (err error) {
	so := &workflow.SessionOptions{
		CreationTimeout:  time.Minute,
		ExecutionTimeout: time.Minute,
	}

	// All activities requested using sessionCtx are guaranteed to execute on the same worker process
	sessionCtx, err := workflow.CreateSession(ctx, so)
	if err != nil {
		return err
	}
	defer workflow.CompleteSession(sessionCtx)

	var downloaded string
	err = workflow.ExecuteActivity(sessionCtx, DownloadFileActivityName, fileName).Get(sessionCtx, &downloaded)
	if err != nil {
		return err
	}

	var processed string
	err = workflow.ExecuteActivity(sessionCtx, ProcessFileActivityName, downloaded).Get(sessionCtx, &processed)
	if err != nil {
		return err
	}

	err = workflow.ExecuteActivity(sessionCtx, UploadFileActivityName, processed).Get(sessionCtx, nil)
	return err
}
