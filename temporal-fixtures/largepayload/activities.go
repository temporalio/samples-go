package largepayload

import (
	"context"
	"crypto/rand"

	"go.temporal.io/sdk/activity"
)

/**
 * Sample activities used by large payloads fixture workflow.
 */

type Activities struct {
}

func (a *Activities) CreateLargeResultActivity(ctx context.Context, sizeBytes int) ([]byte, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Creating large result payload...", sizeBytes)

	token := make([]byte, sizeBytes)
	_, err := rand.Read(token)
	return token, err
}

func (a *Activities) ProcessLargeInputActivity(ctx context.Context, input []byte) error {
	return nil
}
