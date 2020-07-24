package dynamic

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

// SampleGreetingsWorkflow Workflow.
func SampleGreetingsWorkflow(ctx workflow.Context) error {
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
		HeartbeatTimeout:       time.Second * 20,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	var greetResult string
	err := workflow.ExecuteActivity(ctx, "GetGreeting").Get(ctx, &greetResult)
	if err != nil {
		logger.Error("Get greeting failed.", "Error", err)
		return err
	}

	// Get Name.
	var nameResult string
	err = workflow.ExecuteActivity(ctx, "GetName").Get(ctx, &nameResult)
	if err != nil {
		logger.Error("Get name failed.", "Error", err)
		return err
	}

	// Say Greeting.
	var sayResult string
	err = workflow.ExecuteActivity(ctx, "SayGreeting", greetResult, nameResult).Get(ctx, &sayResult)
	if err != nil {
		logger.Error("Marshalling failed with error.", "Error", err)
		return err
	}

	logger.Info("Workflow completed.", "Result", sayResult)
	return nil
}
