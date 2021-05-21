// @@@SNIPSTART samples-go-empty-workflow-definition
package empty

import (
	"go.temporal.io/sdk/workflow"
)

// EmptyWorkflow is a bare minimum Workflow Definition
func EmptyWorkflow(ctx workflow.Context) error {
	return nil
}
// @@@SNIPEND
