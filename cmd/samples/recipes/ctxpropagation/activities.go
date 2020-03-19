package main

import (
	"context"
)

const (
	sampleActivityName = "sampleActivity"
)

func sampleActivity(ctx context.Context) (*Values, error) {
	if val := ctx.Value(propagateKey); val != nil {
		vals := val.(Values)
		return &vals, nil
	}
	return nil, nil
}
