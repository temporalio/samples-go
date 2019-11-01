package main

import (
	"context"

	"go.temporal.io/temporal/activity"
)

const (
	sampleActivityName = "sampleActivity"
)

// This is registration process where you register all your activity handlers.
func init() {
	activity.RegisterWithOptions(
		sampleActivity,
		activity.RegisterOptions{Name: sampleActivityName},
	)
}

func sampleActivity(ctx context.Context) (*Values, error) {
	if val := ctx.Value(propagateKey); val != nil {
		vals := val.(Values)
		return &vals, nil
	}
	return nil, nil
}
