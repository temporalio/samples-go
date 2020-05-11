package childworkflow

import (
	"fmt"
	"time"

	"go.temporal.io/temporal/workflow"
	"go.uber.org/zap"
)

/**
 * This sample workflow demonstrates how to use invoke child workflow from parent workflow execution.  Each child
 * workflow execution is starting a new run and parent execution is notified only after the completion of last run.
 */

// SampleParentWorkflow workflow decider
func SampleParentWorkflow(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)
	execution := workflow.GetInfo(ctx).WorkflowExecution
	// Parent workflow can choose to specify it's own ID for child execution.  Make sure they are unique for each execution.
	childID := fmt.Sprintf("child_workflow:%v", execution.RunID)
	cwo := workflow.ChildWorkflowOptions{
		// Do not specify WorkflowID if you want Temporal server to generate a unique ID for child execution
		WorkflowID:                   childID,
		ExecutionStartToCloseTimeout: time.Minute,
	}
	ctx = workflow.WithChildOptions(ctx, cwo)
	var result string
	childFuture := workflow.ExecuteChildWorkflow(ctx, SampleChildWorkflow)

	// wait for child to start
	var childWE workflow.Execution
	if err := childFuture.GetChildWorkflowExecution().Get(ctx, &childWE); err != nil {
		logger.Error("child execution failed to start.", zap.Error(err))
		return err
	}

	// send signal to child
	signalFuture1 := childFuture.SignalChildWorkflow(ctx, "signal_child", "Hello")
	err := signalFuture1.Get(ctx, nil)
	if err != nil {
		logger.Error("failed to send signal to child.", zap.Error(err))
		return err
	}

	// Receive signal from child
	signalCh := workflow.GetSignalChannel(ctx, "signal_parent")
	var dataFromChild string
	signalCh.Receive(ctx, &dataFromChild)
	logger.Info("Received signal from child", zap.String("data", dataFromChild))

	// wait for child to complete
	err = childFuture.Get(ctx, &result)
	if err != nil {
		logger.Error("Parent execution received child execution failure.", zap.Error(err))
		return err
	}

	logger.Info("Parent execution completed.", zap.String("Result", result))
	return nil
}
