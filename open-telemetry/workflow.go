package open_telemetry

import (
	"context"
	"time"

	"go.temporal.io/sdk/workflow"
)

func Workflow(ctx workflow.Context) error {
	cwo := workflow.ChildWorkflowOptions{
		WorkflowID: "OPEN-TELEMETRY-SIMPLE-CHILD-WORKFLOW-ID",
	}
	ctx = workflow.WithChildOptions(ctx, cwo)
	return workflow.ExecuteChildWorkflow(ctx, ChildWorkflow).Get(ctx, nil)
}

func ChildWorkflow(ctx workflow.Context) error {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	return workflow.ExecuteActivity(ctx, Activity, time.Second).Get(ctx, nil)
}

func Activity(ctx context.Context, d time.Duration) error {
	time.Sleep(d)
	return nil
}
