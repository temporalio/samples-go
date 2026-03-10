package earlyreturn

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/pborman/uuid"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/workflow"
)

const (
	UpdateName    = "early-return"
	TaskQueueName = "early-return-tq"
)

type TransactionRequest struct {
	SourceAccount string
	TargetAccount string
	Amount        int // in cents
}

type Transaction struct {
	ID     string
	Status string
}

// Workflow processes a transaction in two phases. First, the transaction is initialized, and if successful,
// it proceeds to completion. However, if initialization fails - due to validation errors or transient
// issues (e.g., network connectivity problems) - the transaction is cancelled.
//
// By utilizing Update-with-Start, the client can initiate the workflow and immediately receive the result of
// the initialization in a single round trip, even before the transaction processing completes. The remainder
// of the transaction is then processed asynchronously.
func Workflow(ctx workflow.Context, tx TransactionRequest) (*Transaction, error) {
	return run(ctx, tx)
}

func run(ctx workflow.Context, txRequest TransactionRequest) (*Transaction, error) {
	logger := workflow.GetLogger(ctx)

	var tx *Transaction
	var initDone bool
	var initErr error

	if err := workflow.SetUpdateHandler(
		ctx,
		UpdateName,
		func(ctx workflow.Context) (*Transaction, error) {
			if err := workflow.Await(ctx, func() bool { return initDone }); err != nil {
				return nil, fmt.Errorf("transaction init cancelled: %w", err)
			}
			return tx, initErr
		},
	); err != nil {
		return nil, err
	}

	// Phase 1: Initialize the transaction synchronously.
	//
	// By using a local activity, an additional server roundtrip is avoided.
	// See https://docs.temporal.io/activities#local-activity for more details.

	activityOptions := workflow.WithLocalActivityOptions(ctx, workflow.LocalActivityOptions{
		ScheduleToCloseTimeout: 5 * time.Second, // short timeout to avoid another Workflow Task being scheduled
	})

	initErr = workflow.ExecuteLocalActivity(activityOptions, txRequest.Init).Get(ctx, &tx)
	initDone = true

	// Phase 2: Complete or cancel the transaction asychronously.

	activityCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
	})

	if initErr != nil {
		logger.Error(fmt.Sprintf("cancelling transaction due to init error: %v", initErr))

		// Transaction failed to be initialized or not quickly enough; cancel the transaction.
		if err := workflow.ExecuteActivity(activityCtx, CancelTransaction, tx).Get(ctx, nil); err != nil {
			return nil, fmt.Errorf("cancelling the transaction failed: %w", err)
		}

		return nil, initErr
	}

	logger.Info("completing transaction")

	// Transaction was initialized successfully; complete the transaction.
	if err := workflow.ExecuteActivity(activityCtx, CompleteTransaction, tx).Get(ctx, nil); err != nil {
		return nil, fmt.Errorf("completing the transaction failed: %w", err)
	}

	workflow.GetLogger(ctx).Info("Transaction completed successfully", "ID", tx.ID)

	return tx, nil
}

func (tx *TransactionRequest) Init(ctx context.Context) (*Transaction, error) {
	logger := activity.GetLogger(ctx)
	if tx.Amount <= 0 {
		return nil, errors.New("invalid Amount")
	}
	time.Sleep(500 * time.Millisecond)
	logger.Info("Transaction initialized")
	return &Transaction{ID: uuid.New(), Status: "initialized"}, nil
}

func CancelTransaction(ctx context.Context, tx *Transaction) error {
	logger := activity.GetLogger(ctx)
	time.Sleep(1 * time.Second)
	logger.Info("Transaction cancelled")
	return nil
}

func CompleteTransaction(ctx context.Context, tx *Transaction) error {
	logger := activity.GetLogger(ctx)
	time.Sleep(1 * time.Second)
	logger.Info("Transaction completed")
	tx.Status = "completed"
	return nil
}
