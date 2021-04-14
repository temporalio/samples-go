package ctxpropagation

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

// CtxPropWorkflow workflow definition
func CtxPropWorkflow(ctx workflow.Context) (err error) {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 2 * time.Second, // such a short timeout to make sample fail over very fast
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	if val := ctx.Value(PropagateKey); val != nil {
		vals := val.(Values)
		workflow.GetLogger(ctx).Info("custom context propagated to workflow", vals.Key, vals.Value)
	}

	var values Values
	if err = workflow.ExecuteActivity(ctx, SampleActivity).Get(ctx, &values); err != nil {
		workflow.GetLogger(ctx).Error("Workflow failed.", "Error", err)
		return err
	}
	workflow.GetLogger(ctx).Info("context propagated to activity", values.Key, values.Value)
	workflow.GetLogger(ctx).Info("Workflow completed.")
	return nil
}
