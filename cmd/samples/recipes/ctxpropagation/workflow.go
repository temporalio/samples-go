package main

import (
	"time"

	"go.temporal.io/temporal/workflow"
	"go.uber.org/zap"
)

// ApplicationName is the task list for this sample
const ApplicationName = "CtxPropagatorGroup"

// This is registration process where you register all your workflow handlers.
func init() {
	workflow.Register(CtxPropWorkflow)
}

// CtxPropWorkflow workflow decider
func CtxPropWorkflow(ctx workflow.Context) (err error) {
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Second * 5,
		StartToCloseTimeout:    time.Minute,
		HeartbeatTimeout:       time.Second * 2, // such a short timeout to make sample fail over very fast
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	if val := ctx.Value(propagateKey); val != nil {
		vals := val.(Values)
		workflow.GetLogger(ctx).Info("custom context propagated to workflow", zap.String(vals.Key, vals.Value))
	}

	var values Values
	if err = workflow.ExecuteActivity(ctx, sampleActivity).Get(ctx, &values); err != nil {
		workflow.GetLogger(ctx).Error("Workflow failed.", zap.Error(err))
		return err
	}
	workflow.GetLogger(ctx).Info("context propagated to activity", zap.String(values.Key, values.Value))
	workflow.GetLogger(ctx).Info("Workflow completed.")
	return nil
}
