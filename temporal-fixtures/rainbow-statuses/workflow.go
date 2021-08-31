package rainbowstatuses

import (
	"time"

	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// RainbowStatusesWorkflow workflow definition.
func RainbowStatusesWorkflow(ctx workflow.Context, status enums.WorkflowExecutionStatus) error {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 1 * time.Hour,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 1,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("RainbowStatusesWorkflow workflow started", "Status", status)

	var a *Activities
	var err error
	switch status {
	case enums.WORKFLOW_EXECUTION_STATUS_COMPLETED:
		err = workflow.ExecuteActivity(ctx, a.CompletedActivity).Get(ctx, nil)
	case enums.WORKFLOW_EXECUTION_STATUS_RUNNING:
		err = workflow.ExecuteActivity(ctx, a.LongActivity).Get(ctx, nil)
	case enums.WORKFLOW_EXECUTION_STATUS_TERMINATED:
		err = workflow.ExecuteActivity(ctx, a.LongActivity).Get(ctx, nil)
	case enums.WORKFLOW_EXECUTION_STATUS_CANCELED:
		err = workflow.ExecuteActivity(ctx, a.LongActivity).Get(ctx, nil)
	case enums.WORKFLOW_EXECUTION_STATUS_FAILED:
		err = workflow.ExecuteActivity(ctx, a.FailedActivity).Get(ctx, nil)
	case enums.WORKFLOW_EXECUTION_STATUS_TIMED_OUT:
		err = workflow.ExecuteActivity(ctx, a.LongActivity).Get(ctx, nil)
	case enums.WORKFLOW_EXECUTION_STATUS_CONTINUED_AS_NEW:
		return workflow.NewContinueAsNewError(ctx, RainbowStatusesWorkflow, enums.WORKFLOW_EXECUTION_STATUS_COMPLETED)
	}

	if err != nil {
		logger.Error("Activity failed.", "Error", err)
		return err
	}

	logger.Info("RainbowStatusesWorkflow workflow completed.")

	return nil
}
