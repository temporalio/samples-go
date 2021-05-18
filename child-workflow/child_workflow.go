// @@@SNIPSTART samples-go-child-workflow-example-child-workflow-type
package child_workflow

import (
	"go.temporal.io/sdk/workflow"
)

// SampleChildWorkflow is a Workflow Type
func SampleChildWorkflow(ctx workflow.Context, name string) (string, error) {
	logger := workflow.GetLogger(ctx)
	greeting := "Hello " + name + "!"
	logger.Info("Child workflow execution: " + greeting)
	return greeting, nil
}
// @@@SNIPEND
