// @@@START go-helloworld-sample-activity
package helloworld

import (
  "context"

  "go.temporal.io/sdk/activity"
)

func HelloWorldActivity(ctx context.Context, name string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Activity is executing\n")
  // Append the name to the greeting and return it
  greeting := "Hello " + name + "!"
	return greeting, nil
}
// @@@END"
