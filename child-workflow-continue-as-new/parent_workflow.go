package child_workflow_continue_as_new

import (
	"fmt"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

// SampleParentWorkflow workflow definition
func SampleParentWorkflow(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)
	execution := workflow.GetInfo(ctx).WorkflowExecution
	// Parent workflow can choose to specify it's own ID for child execution.  Make sure they are unique for each execution.
	childID := fmt.Sprintf("child_workflow:%v", execution.RunID)
	cwo := workflow.ChildWorkflowOptions{
		// Do not specify WorkflowID if you want Temporal server to generate a unique ID for child execution
		WorkflowID: childID,
	}
	ctx = workflow.WithChildOptions(ctx, cwo)
	var result string
	err := workflow.ExecuteChildWorkflow(ctx, SampleChildWorkflow, 0, 5).Get(ctx, &result)
	if err != nil {
		logger.Error("Parent execution received child execution failure.", zap.Error(err))
		return err
	}

	logger.Info("Parent execution completed.", zap.String("Result", result))
	return nil
}
