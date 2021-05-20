package largepayload

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

// LargePayloadWorkflow workflow definition
func LargePayloadWorkflow(ctx workflow.Context, payloadSize int) (err error) {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var data []byte
	var a *Activities
	err = workflow.ExecuteActivity(ctx, a.CreateLargeResultActivity, payloadSize).Get(ctx, &data)
	if err != nil {
		return err
	}

	err = workflow.ExecuteActivity(ctx, a.ProcessLargeInputActivity, data).Get(ctx, nil)

	if err != nil {
		workflow.GetLogger(ctx).Error("Workflow failed.", "Error", err.Error())
	} else {
		workflow.GetLogger(ctx).Info("Workflow completed.")
	}
	return err
}
