package earlyreturn

import (
	"errors"
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"
)

const (
	UpdateName         = "early-return"
	TaskQueueName      = "early-return-tq"
	activityTimeout    = 2 * time.Second
	earlyReturnTimeout = 5 * time.Second
)

type Transaction struct {
	ID          string
	FromAccount string
	ToAccount   string
	Amount      float64
}

// Workflow processes a transaction in two phases. First, the transaction is initialized, and if successful,
// it proceeds to completion. However, if initialization fails - due to validation errors or transient
// issues (e.g., network connectivity problems) - the transaction is cancelled.
//
// By utilizing Update-with-Start, the client can initiate the workflow and immediately receive the result of
// the initialization in a single round trip, even before the transaction processing completes. The remainder
// of the transaction is then processed asynchronously.
func Workflow(ctx workflow.Context, tx Transaction) error {
	var initErr error
	var initDone bool
	logger := workflow.GetLogger(ctx)

	if err := workflow.SetUpdateHandler(
		ctx,
		UpdateName,
		func(ctx workflow.Context) error {
			condition := func() bool { return initDone }
			if completed, err := workflow.AwaitWithTimeout(ctx, earlyReturnTimeout, condition); err != nil {
				return fmt.Errorf("update cancelled: %w", err)
			} else if !completed {
				return errors.New("update timed out")
			}
			return initErr
		},
	); err != nil {
		return err
	}

	// Phase 1: Initialize the transaction synchronously.
	//
	// By using a local activity, an additional server roundtrip is avoided.
	// See https://docs.temporal.io/activities#local-activity for more details.

	activityOptions := workflow.WithLocalActivityOptions(ctx, workflow.LocalActivityOptions{
		ScheduleToCloseTimeout: activityTimeout,
	})
	initErr = workflow.ExecuteLocalActivity(activityOptions, InitTransaction, tx).Get(ctx, nil)
	initDone = true

	// Phase 2: Complete or cancel the transaction asychronously.
	activityCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	})
	if initErr != nil {
		logger.Info("cancelling transaction due to error: %v", initErr)

		// Transaction failed to be initialized or not quickly enough; cancel the transaction.
		return workflow.ExecuteActivity(activityCtx, CancelTransaction, tx).Get(ctx, nil)
	}

	logger.Info("completing transaction")

	// Transaction was initialized successfully; complete the transaction.
	return workflow.ExecuteActivity(activityCtx, CompleteTransaction, tx).Get(ctx, nil)
}
