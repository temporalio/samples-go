package batch_sliding_window

import (
	"go.temporal.io/sdk/workflow"
	"math/rand"
	"time"
)

// RecordProcessorWorkflow workflow that implements processing of a single record.
func RecordProcessorWorkflow(ctx workflow.Context, r SingleRecord) error {
	err := ProcessRecord(ctx, r)
	// Notify parent about completion via signal
	parent := workflow.GetInfo(ctx).ParentWorkflowExecution
	// This workflow is always expected to have a parent.
	// But for unit testing it might be useful to skip the notification if there is none.
	if parent != nil {
		// Doesn't specify runId as parent calls continue-as-new.
		signaled := workflow.SignalExternalWorkflow(ctx, parent.ID, "", "ReportCompletion", r.Id)
		// Ensure that signal is delivered.
		// Completing workflow before this Future is ready might lead to the signal loss.
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

	// Use SideEffect to get a random number to ensure workflow determinism.
	encodedRandom := workflow.SideEffect(ctx, func(ctx workflow.Context) interface{} {
		//workflowcheck:ignore
		return rand.Intn(10)
	})
	var random int
	err := encodedRandom.Get(&random)
	if err != nil {
		return err
	}
	err = workflow.Sleep(ctx, time.Duration(random)*time.Second)
	if err != nil {
		return err
	}
	workflow.GetLogger(ctx).Info("Processed ", r)
	return nil
}
