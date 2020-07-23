package cancelactivity

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"
)

/**
 * This is the cancel activity workflow sample.
 */

// Workflow workflow
func Workflow(ctx workflow.Context) error {
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute * 30,
		HeartbeatTimeout:       time.Second * 5,
		WaitForCancellation:    true,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	logger := workflow.GetLogger(ctx)
	logger.Info("cancel workflow started")
	var a *Activities // Used to call activities by function pointer
	defer func() {
		// When workflow is canceled, it has to get a new disconnected context to execute any activities
		newCtx, _ := workflow.NewDisconnectedContext(ctx)
		err := workflow.ExecuteActivity(newCtx, a.CleanupActivity).Get(ctx, nil)
		if err != nil {
			logger.Error("Cleanup activity failed", "Error", err)
		}
	}()

	var result string
	err := workflow.ExecuteActivity(ctx, a.ActivityToBeCanceled).Get(ctx, &result)
	logger.Info(fmt.Sprintf("activityToBeCanceled returns %v, %v", result, err))

	err = workflow.ExecuteActivity(ctx, a.ActivityToBeSkipped).Get(ctx, nil)
	logger.Error("Error from activityToBeSkipped", "Error", err)

	logger.Info("Workflow completed.")

	return nil
}
