package child_workflow

import (
	"go.temporal.io/temporal/workflow"
)

// SampleChildWorkflow workflow definition
func SampleChildWorkflow(ctx workflow.Context, name string) (string, error) {
	logger := workflow.GetLogger(ctx)
	greeting := "Hello " + name + "!"
	logger.Info("Child workflow execution: " + greeting)
	return greeting, nil
}
