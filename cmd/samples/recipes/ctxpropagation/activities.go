package main

import (
	"context"

	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/workflow"
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

func sampleActivity(ctx context.Context) (map[string]string, error) {
	values := make(map[string]string, len(propagatedKeys))
	for _, key := range propagatedKeys {
		if val, ok := ctx.Value(workflow.ContextKey(key)).(string); ok {
			values[key] = val
		}
	}
	return values, nil
}
