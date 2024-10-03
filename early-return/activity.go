package earlyreturn

import (
	"context"
	"errors"

	"go.temporal.io/sdk/activity"
)

func InitTransaction(ctx context.Context, tx Transaction) error {
	logger := activity.GetLogger(ctx)
	if tx.FromAccount == "" {
		return errors.New("invalid fromAccount")
	}
	if tx.ToAccount == "" {
		return errors.New("invalid toAccount")
	}
	if tx.Amount == 0 {
		return errors.New("invalid amount")
	}
	logger.Info("Transaction initialized")
	return nil
}

func CancelTransaction(ctx context.Context, tx Transaction) {
	logger := activity.GetLogger(ctx)
	logger.Info("Transaction cancelled")
}

func CompleteTransaction(ctx context.Context, tx Transaction) {
	logger := activity.GetLogger(ctx)
	logger.Info("Transaction completed")
}
