package batch_sliding_window

import (
	"go.temporal.io/sdk/workflow"
	"time"
)

func RecordProcessorWorkflow(ctx workflow.Context, r SingleRecord) error {
	err := ProcessRecord(ctx, r)
	// Notify parent about completion via signal
	parent := workflow.GetInfo(ctx).ParentWorkflowExecution
	// This workflow is always expected to have a parent.
	// But for unit testing it might be useful to skip the notification if there is none.
	if parent != nil {
		// Doesn't specify runId as parent calls continue-as-new.
		signaled := workflow.SignalExternalWorkflow(ctx, parent.ID, "", "ReportCompletion", r.Id)
		signalErr := signaled.Get(ctx, nil)
		if signalErr != nil {
			return signalErr
		}
	}
	return err
}

// ProcessRecord simulates application specific record processing.
func ProcessRecord(ctx workflow.Context, r SingleRecord) error {
	// Simulate some processing
	workflow.Sleep(ctx, 10*time.Second)
	workflow.GetLogger(ctx).Info("Processed ", r)
	return nil
}
