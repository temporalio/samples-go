package earlyreturn

import (
	"context"
	"errors"

	"go.temporal.io/sdk/activity"
)

func InitTransaction(ctx context.Context, transactionId, fromAccount, toAccount string, amount float64) error {
	logger := activity.GetLogger(ctx)
	if fromAccount == "" {
		return errors.New("invalid fromAccount")
	}
	if toAccount == "" {
		return errors.New("invalid toAccount")
	}
	if amount == 0 {
		return errors.New("invalid amount")
	}
	logger.Info("Transaction initialized")
	return nil
}

func CancelTransaction(ctx context.Context, transactionId string) {
	logger := activity.GetLogger(ctx)
	logger.Info("Transaction cancelled")
}

func CompleteTransaction(ctx context.Context, transactionId string) {
	logger := activity.GetLogger(ctx)
	logger.Info("Transaction completed")
}
