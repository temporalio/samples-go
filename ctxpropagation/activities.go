package ctxpropagation

import (
	"context"
)

// @@@SNIPSTART samples-go-ctx-propagation-activity
func SampleActivity(ctx context.Context) (*Values, error) {
	if val := ctx.Value(PropagateKey); val != nil {
		vals := val.(Values)
		return &vals, nil
	}
	return nil, nil
}
// @@@SNIPEND
