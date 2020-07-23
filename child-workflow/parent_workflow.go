package child_workflow

import (
	"go.temporal.io/sdk/workflow"
)

// This sample workflow demonstrates how to use invoke child workflow from parent workflow execution.  Each child
// workflow execution is starting a new run and parent execution is notified only after the completion of last run.

// SampleParentWorkflow workflow definition
func SampleParentWorkflow(ctx workflow.Context) (string, error) {
	logger := workflow.GetLogger(ctx)

	cwo := workflow.ChildWorkflowOptions{}
	ctx = workflow.WithChildOptions(ctx, cwo)

	var result string
	err := workflow.ExecuteChildWorkflow(ctx, SampleChildWorkflow, "World").Get(ctx, &result)
	if err != nil {
		logger.Error("Parent execution received child execution failure.", "Error", err)
		return "", err
	}
	logger.Info("Parent execution completed.", "Result", result)
	return result, nil
}
