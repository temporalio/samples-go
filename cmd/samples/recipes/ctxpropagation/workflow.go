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
const ApplicationName = "CtxPropagatorGroup"

// HostID - Use a new uuid just for demo so we can run 2 host specific activity workers on same machine.
// In real world case, you would use a hostname or ip address as HostID.
var HostID = ApplicationName + "_" + uuid.New()

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

	for _, key := range propagatedKeys {
		if val, ok := ctx.Value(workflow.ContextKey(key)).(string); ok {
			workflow.GetLogger(ctx).Info("context propagated to workflow", zap.String(key, val))
		}
	}
	vals := ctx.Value(propagateKey).(Values)
	workflow.GetLogger(ctx).Info("custom context propagated to workflow", zap.String(vals.Key, vals.Value))

	var values map[string]string
	if err = workflow.ExecuteActivity(ctx, sampleActivity).Get(ctx, &values); err != nil {
		workflow.GetLogger(ctx).Error("Workflow failed.", zap.Error(err))
		return err
	}
	for key, val := range values {
		workflow.GetLogger(ctx).Info("context propagated to activity", zap.String(key, val))
	}
	workflow.GetLogger(ctx).Info("Workflow completed.")
	return nil
}
