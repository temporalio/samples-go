package largepayload

import (
	"context"
)

/**
 * Sample activities used by large payloads fixture workflow.
 */

type Activities struct {
}

func (a *Activities) Activity(ctx context.Context, input []byte) error {
	return nil
}
