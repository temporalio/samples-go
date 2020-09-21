// @@@SNIPSTART go-helloworld-sample-workflow
package helloworld

import (
  "time"
  
  "go.temporal.io/sdk/workflow"
)

func HelloWorldWorkflow(ctx workflow.Context, name string) (string, error) {
  // Create Activity options
  ao := workflow.ActivityOptions{
    ScheduleToCloseTimeout: time.Minute,
  }
  // Apply them to the Workflow context
  ctx = workflow.WithActivityOptions(ctx, ao)
  // Get a logger to print to the console
  logger := workflow.GetLogger(ctx)
  logger.Info("HelloWorld workflow started\n")
  // Execute the Activity within the Workflow
  var result string
  err := workflow.ExecuteActivity(ctx, HelloWorldActivity, name).Get(ctx, &result)
  if err != nil {
    logger.Error("Activity failed.", "Error", err)
    return "", err
  }
  logger.Info("HelloWorld workflow completed\n")
  // Return the Activity result back to the starter
  return result, nil
}
// @@@SNIPEND
