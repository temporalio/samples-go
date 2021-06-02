package uidriven

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/temporal"
)

func RegisterEmail(ctx context.Context, email string) error {
	logger := activity.GetLogger(ctx)

	logger.Info("activity: registered email", email)

	return nil
}

func ValidateSize(ctx context.Context, size string) error {
	if size != "small" && size != "medium" && size != "large" {
		return temporal.NewNonRetryableApplicationError(
			fmt.Sprintf("size: %s is not valid", size),
			"InvalidSize",
			nil,
			nil,
		)
	}

	return nil
}

func ValidateColor(ctx context.Context, color string) error {
	if color != "red" && color != "blue" {
		return temporal.NewNonRetryableApplicationError(
			fmt.Sprintf("color: %s is not valid", color),
			"InvalidColor",
			nil,
			nil,
		)
	}

	return nil
}

func ProcessOrder(ctx context.Context, order TShirtOrder) error {
	logger := activity.GetLogger(ctx)

	logger.Info("activity: processed order", order)

	return nil
}
