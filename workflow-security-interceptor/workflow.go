package workflow_security_interceptor

import (
	"go.temporal.io/sdk/workflow"
)

func ChildWorkflow(ctx workflow.Context) (string, error) {
	return "OK", nil
}

func ProhibitedChildWorkflow(ctx workflow.Context) (string, error) {
	return "OK", nil
}

func Workflow(ctx workflow.Context) error {
	err := workflow.ExecuteChildWorkflow(ctx, ChildWorkflow).Get(ctx, nil)
	if err != nil {
		return err
	}
	// This will fail because the child workflow type is not allowed
	err = workflow.ExecuteChildWorkflow(ctx, ProhibitedChildWorkflow).Get(ctx, nil)
	if err != nil {
		return err
	}
	return nil
}
