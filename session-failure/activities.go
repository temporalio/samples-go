package fileprocessing

import (
	"context"
	"time"

	"go.temporal.io/sdk/activity"
)

/**
 * Sample activities used by session failure sample workflow.
 */

type Activities struct {
}

func (a *Activities) PrepareWorkerActivity(ctx context.Context) error {
	logger := activity.GetLogger(ctx)
	logger.Info("Preparing session worker")
	return nil
}

func (a *Activities) LongRunningActivity(ctx context.Context) error {
	logger := activity.GetLogger(ctx)
	logger.Info("Started running long running activity.")

	hbTicker := time.NewTicker(20 * time.Second)
	defer hbTicker.Stop()
	// Create a 5 minute timer to simulate an activity doing some long work
	timer := time.NewTimer(5 * time.Minute)
	defer timer.Stop()
	for {
		select {
		case <-hbTicker.C:
			activity.RecordHeartbeat(ctx)
		case <-timer.C:
			return ctx.Err()
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
