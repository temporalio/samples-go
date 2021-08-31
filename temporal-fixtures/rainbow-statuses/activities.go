package rainbowstatuses

import (
	"context"
	"errors"
	"time"
)

/**
 * Sample activities used by large payloads fixture workflow.
 */

type Activities struct {
}

func (a *Activities) CompletedActivity(ctx context.Context) error {
	return nil
}

func (a *Activities) FailedActivity(ctx context.Context) error {
	return errors.New("manual failure")
}

func (a *Activities) LongActivity(ctx context.Context) error {
	time.Sleep(24 * time.Hour)
	return nil
}
