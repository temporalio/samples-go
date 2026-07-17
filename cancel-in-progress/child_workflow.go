package cancel_in_progress

import (
	"go.temporal.io/sdk/workflow"
	"time"
)

// SampleChildWorkflow is a Workflow Definition
func SampleChildWorkflow(ctx workflow.Context, name string) (string, error) {
	logger := workflow.GetLogger(ctx)

	// Simulate some long running processing.
	_ = workflow.Sleep(ctx, time.Second*3)

	greeting := "Hello " + name + "!"
	logger.Info("Child workflow execution: " + greeting)

	return greeting, nil
}
