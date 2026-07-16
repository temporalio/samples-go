package datadog

import (
	"context"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/workflow"
)

func Workflow(ctx workflow.Context, name string) error {
	workflow.GetLogger(ctx).Info("Executing Workflow.", "name", name)
	cwo := workflow.ChildWorkflowOptions{
		WorkflowID: "DATADOG-CHILD-WORKFLOW-ID",
	}
	ctx = workflow.WithChildOptions(ctx, cwo)
	return workflow.ExecuteChildWorkflow(ctx, ChildWorkflow, name).Get(ctx, nil)
}

func ChildWorkflow(ctx workflow.Context, name string) error {
	workflow.GetLogger(ctx).Info("Executing ChildWorkflow.", "name", name)
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	return workflow.ExecuteActivity(ctx, Activity, name).Get(ctx, nil)
}

func Activity(ctx context.Context, name string) error {
	activity.GetLogger(ctx).Info("Executing Activity.", "name", name)
	time.Sleep(time.Second)
	return nil
}
