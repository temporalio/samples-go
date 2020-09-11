// @@@START go-helloworld-sample-workflow
package helloworld

import (
  "time"

  "go.temporal.io/sdk/workflow"
)

func HelloWorldWorkflow(ctx workflow.Context, name string) (string, error) {
  ao := workflow.ActivityOptions{
		ScheduleToCloseTimeout: time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("HelloWorld workflow started\n")

	var result string
	err := workflow.ExecuteActivity(ctx, HelloWorldActivity, name).Get(ctx, &result)
	if err != nil {
		logger.Error("Activity failed.", "Error", err)
		return "", err
	}

  logger.Info("HelloWorld workflow completed\n")

  return result, nil
}
// @@@END
