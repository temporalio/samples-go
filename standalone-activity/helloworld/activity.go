package helloworld

import (
	"context"
	"go.temporal.io/sdk/activity"
)

func Activity(ctx context.Context, name string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Activity", "name", name)
	return "Hello " + name + "!", nil
}
