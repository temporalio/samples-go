package main

import (
	"context"
	"time"

	"go.uber.org/cadence"
	"go.uber.org/zap"
)

/**
 * This is the hello world workflow sample.
 */

// ApplicationName is the task list for this sample
const ApplicationName = "helloWorldGroup"

// This is registration process where you register all your workflows
// and activity function handlers.
func init() {
	cadence.RegisterWorkflow(Workflow)
	cadence.RegisterActivity(helloworldActivity)
}

// Workflow workflow decider
func Workflow(ctx cadence.Context, name string) error {
	ao := cadence.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
		HeartbeatTimeout:       time.Second * 20,
	}
	ctx = cadence.WithActivityOptions(ctx, ao)

	logger := cadence.GetLogger(ctx)
	logger.Info("helloworld workflow started")
	var helloworldResult string
	err := cadence.ExecuteActivity(ctx, helloworldActivity, name).Get(ctx, &helloworldResult)
	if err != nil {
		logger.Error("Activity failed.", zap.Error(err))
		return err
	}

	logger.Info("Workflow completed.", zap.String("Result", helloworldResult))

	return nil
}

func helloworldActivity(ctx context.Context, name string) (string, error) {
	logger := cadence.GetActivityLogger(ctx)
	logger.Info("helloworld activity started")
	return "Hello " + name + "!", nil
}
