package main

import (
	"context"
	"fmt"

	"go.uber.org/cadence/activity"
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

func sampleActivity(ctx context.Context) error {
	if val := ctx.Value(propagateKey); val != nil {
		vals := val.(Values)
		fmt.Printf("custom context propagated to activity %v\n", vals)
	}
	return nil
}
