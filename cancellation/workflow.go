package cancellation

import (
	"errors"
	"go.temporal.io/sdk/workflow"
	"time"
)

// @@@SNIPSTART samples-go-cancellation-workflow-definition
// YourWorkflow is a Workflow Definition that shows how it can be canceled.
func YourWorkflow(ctx workflow.Context) error {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Minute,
		HeartbeatTimeout:    5 * time.Second,
		WaitForCancellation: true,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	logger := workflow.GetLogger(ctx)
	logger.Info("cancel workflow started")
	var a *Activities // Used to call Activities by function pointer
	defer func() {

		if !errors.Is(ctx.Err(), workflow.ErrCanceled) {
			return
		}

		// When the Workflow is canceled, it has to get a new disconnected context to execute any Activities
		// Specify activity options for cleanup activity
		aoCleanup := workflow.ActivityOptions{
			StartToCloseTimeout: 2 * time.Second,
		}
		newCtx, _ := workflow.NewDisconnectedContext(ctx)
		newCtx = workflow.WithActivityOptions(newCtx, aoCleanup)
		err := workflow.ExecuteActivity(newCtx, a.CleanupActivity).Get(ctx, nil)
		if err != nil {
			logger.Error("CleanupActivity failed", "Error", err)
		}

	}()

	var result string
	err := workflow.ExecuteActivity(ctx, a.ActivityToBeCanceled).Get(ctx, &result)
	if err != nil {
		return err
	}

	logger.Info("Workflow Execution complete.")
	return nil
}
