package child_workflow

import (
	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/workflow"
)

// @@@SNIPSTART samples-go-child-workflow-example-parent-workflow-definition
// SampleParentWorkflow is a Workflow Definition
// This Workflow Definition demonstrates how to start a Child Workflow Execution from a Parent Workflow Execution.
// Each Child Workflow Execution starts a new Run.
// The Parent Workflow Execution is notified only after the completion of last Run of the Child Workflow Execution.
func SampleParentWorkflow(ctx workflow.Context) (string, error) {
	logger := workflow.GetLogger(ctx)

	cwo := workflow.ChildWorkflowOptions{
		WorkflowID:            "ABC-SIMPLE-CHILD-WORKFLOW-ID",
		WorkflowIDReusePolicy: enumspb.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
	}
	ctx = workflow.WithChildOptions(ctx, cwo)

	var result string
	err := workflow.ExecuteChildWorkflow(ctx, SampleChildWorkflow, "World").GetChildWorkflowExecution().Get(ctx, nil)
	if err != nil {
		logger.Error("Parent execution received child execution failure.", "Error", err)
		return "", err
	}

	err = workflow.ExecuteChildWorkflow(ctx, SampleChildWorkflow, "Foo").Get(ctx, &result)
	if err != nil {
		logger.Error("Parent execution received child execution failure.", "Error", err)
		return "", err
	}

	logger.Info("Parent execution completed.", "Result", result)
	return result, nil
}

// @@@SNIPEND
