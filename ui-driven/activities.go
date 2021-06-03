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
	for _, key := range TShirtSizes {
		if key == size {
			return nil
		}
	}

	return temporal.NewNonRetryableApplicationError(
		fmt.Sprintf("size: %s is not valid (%v)", size, TShirtSizes),
		"InvalidSize",
		nil,
		nil,
	)
}

func ValidateColor(ctx context.Context, color string) error {
	for _, key := range TShirtColors {
		if key == color {
			return nil
		}
	}

	return temporal.NewNonRetryableApplicationError(
		fmt.Sprintf("color: %s is not valid (%v)", color, TShirtColors),
		"InvalidColor",
		nil,
		nil,
	)
}

func ProcessOrder(ctx context.Context, order TShirtOrder) error {
	logger := activity.GetLogger(ctx)

	logger.Info("activity: processed order", order)

	return nil
}
