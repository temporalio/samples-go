package childworkflow

import (
	"go.temporal.io/temporal/workflow"
	"go.uber.org/zap"
)

/**
 * This sample workflow demonstrates how to use invoke child workflow from parent workflow execution.  Each child
 * workflow execution is starting a new run and parent execution is notified only after the completion of last run.
 */

// SampleChildWorkflow workflow decider
func SampleChildWorkflow(ctx workflow.Context) (string, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Child workflow execution started.")

	// Receive signal from parent
	signalCh := workflow.GetSignalChannel(ctx, "signal_child")
	var dataFromParent string
	signalCh.Receive(ctx, &dataFromParent)
	logger.Info("Received signal from parent", zap.String("data", dataFromParent))

	// Send signal to parent
	parentWF := workflow.GetInfo(ctx).ParentWorkflowExecution
	signalFuture := workflow.SignalExternalWorkflow(ctx, parentWF.ID, parentWF.RunID, "signal_parent", "World")
	err := signalFuture.Get(ctx, nil)
	if err != nil {
		logger.Error("failed to send signal to parent.", zap.Error(err))
		return "", err
	}

	logger.Info("Child execution completed.", zap.String("Result", dataFromParent))
	return "done", nil
}
