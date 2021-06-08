package child_workflow_continue_as_new

import (
	"fmt"

	"go.temporal.io/sdk/workflow"
)
// @@@SNIPSTART samples-go-cw-cas-parent-workflow-definition
// SampleParentWorkflow is a Workflow Definition
func SampleParentWorkflow(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)
	execution := workflow.GetInfo(ctx).WorkflowExecution
	// Parent Workflows can choose to specify Ids for child executions.
	// Make sure Ids are unique for each execution.
	// Do not specify if you want the Temporal Server to generate a unique ID for the child execution.
	childID := fmt.Sprintf("child_workflow:%v", execution.RunID)
	cwo := workflow.ChildWorkflowOptions{
		WorkflowID: childID,
	}
	ctx = workflow.WithChildOptions(ctx, cwo)
	var result string
	err := workflow.ExecuteChildWorkflow(ctx, SampleChildWorkflow, 0, 5).Get(ctx, &result)
	if err != nil {
		logger.Error("Parent execution received child execution failure.", "Error", err)
		return err
	}

	logger.Info("Parent execution completed.", "Result", result)
	return nil
}
// @@@SNIPEND
