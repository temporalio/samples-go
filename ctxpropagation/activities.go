package ctxpropagation

import (
	"context"
)

func SampleActivity(ctx context.Context) (*Values, error) {
	if val := ctx.Value(PropagateKey); val != nil {
		vals := val.(Values)
		return &vals, nil
	}
	return nil, nil
}
