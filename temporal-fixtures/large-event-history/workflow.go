package largepayload

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

// LargePayloadWorkflow workflow definition
func Workflow(ctx workflow.Context, LengthOfHistory int, WillFailOrNot bool) (err error) {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var data []byte
	var a *Activities
	i := 1
	for i <= LengthOfHistory {
		err = workflow.ExecuteActivity(ctx, a.Activity, data).Get(ctx, nil)
	}

	if err != nil {
		workflow.GetLogger(ctx).Error("Workflow failed.", "Error", err.Error())
	} else {
		workflow.GetLogger(ctx).Info("Workflow completed.")
	}
	return err
}
