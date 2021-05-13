package fileprocessing

import (
	"time"

	"go.temporal.io/sdk/temporal"

	"go.temporal.io/sdk/workflow"
)

// SampleFileProcessingWorkflow workflow definition
func SampleFileProcessingWorkflow(ctx workflow.Context, fileName string) (err error) {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
		HeartbeatTimeout:    2 * time.Second, // such a short timeout to make sample fail over very fast
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	// Retry the whole sequence from the first activity on any error
	// to retry it on a different host. In a real application it might be reasonable to
	// retry individual activities as well as the whole sequence discriminating between different types of errors.
	// See the retryactivity sample for a more sophisticated retry implementation.
	for i := 1; i < 5; i++ {
		err = processFile(ctx, fileName)
		if err == nil {
			break
		}
	}
	if err != nil {
		workflow.GetLogger(ctx).Error("Workflow failed.", "Error", err.Error())
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
	sessionCtx, err := workflow.CreateSession(ctx, so)
	if err != nil {
		return err
	}
	defer workflow.CompleteSession(sessionCtx)

	var downloadedName string
	var a *Activities
	err = workflow.ExecuteActivity(sessionCtx, a.DownloadFileActivity, fileName).Get(sessionCtx, &downloadedName)
	if err != nil {
		return err
	}

	var processedFileName string
	err = workflow.ExecuteActivity(sessionCtx, a.ProcessFileActivity, downloadedName).Get(sessionCtx, &processedFileName)
	if err != nil {
		return err
	}

	err = workflow.ExecuteActivity(sessionCtx, a.UploadFileActivity, processedFileName).Get(sessionCtx, nil)
	return err
}
