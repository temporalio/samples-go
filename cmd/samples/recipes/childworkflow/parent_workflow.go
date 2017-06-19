package main

import (
	"fmt"
	"time"

	"go.uber.org/cadence"
	"go.uber.org/zap"
)

/**
 * This sample workflow demonstrates how to use invoke child workflow from parent workflow execution.  Each child
 * workflow execution is starting a new run and parent execution is notified only after the completion of last run.
 */

// ApplicationName is the task list for this sample
const ApplicationName = "childWorkflowGroup"

// This is registration process where you register all your workflows
// and activity function handlers.
func init() {
	cadence.RegisterWorkflow(SampleParentWorkflow)
}

// SampleParentWorkflow workflow decider
func SampleParentWorkflow(ctx cadence.Context) error {
	logger := cadence.GetLogger(ctx)
	execution := cadence.GetWorkflowInfo(ctx).WorkflowExecution
	// Parent workflow can choose to specify it's own ID for child execution.  Make sure they are unique for each execution.
	childID := fmt.Sprintf("child_workflow:%v", execution.RunID)
	cwo := cadence.ChildWorkflowOptions{
		// Do not specify WorkflowID if you want cadence to generate a unique ID for child execution
		WorkflowID:                   childID,
		ExecutionStartToCloseTimeout: time.Minute,
	}
	ctx = cadence.WithChildWorkflowOptions(ctx, cwo)
	var result string
	err := cadence.ExecuteChildWorkflow(ctx, SampleChildWorkflow, 0, 5).Get(ctx, &result)
	if err != nil {
		logger.Error("Parent execution received child execution failure.", zap.Error(err))
		return err
	}

	logger.Info("Parent execution completed.", zap.String("Result", result))
	return nil
}
