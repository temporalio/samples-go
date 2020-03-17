package main

import (
	"fmt"
	"time"

	"go.temporal.io/temporal/workflow"
	"go.uber.org/zap"
)

/**
 * The purpose of this sample is to demonstrate invocation of workflows and activities using name rather than strongly
 * typed function.
 */

// ApplicationName is the task list for this sample
const ApplicationName = "dynamicGroup"

// GreetingsWorkflowName name used when workflow function is registered during init.  We use the fully qualified name to function
const GreetingsWorkflowName = "SampleGreetingsWorkflow"

// Activity names used when activity function is registered during init.  We use the fully qualified name to function
const getNameActivityName = "getNameActivity"
const getGreetingActivityName = "getGreetingActivity"
const sayGreetingActivityName = "sayGreetingActivity"

// SampleGreetingsWorkflow Workflow Decider.
func SampleGreetingsWorkflow(ctx workflow.Context) error {
	// Get Greeting.
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
		HeartbeatTimeout:       time.Second * 20,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	var greetResult string
	err := workflow.ExecuteActivity(ctx, getGreetingActivityName).Get(ctx, &greetResult)
	if err != nil {
		logger.Error("Get greeting failed.", zap.Error(err))
		return err
	}

	// Get Name.
	var nameResult string
	err = workflow.ExecuteActivity(ctx, getNameActivityName).Get(ctx, &nameResult)
	if err != nil {
		logger.Error("Get name failed.", zap.Error(err))
		return err
	}

	// Say Greeting.
	var sayResult string
	err = workflow.ExecuteActivity(ctx, sayGreetingActivityName, greetResult, nameResult).Get(ctx, &sayResult)
	if err != nil {
		logger.Error("Marshalling failed with error.", zap.Error(err))
		return err
	}

	logger.Info("Workflow completed.", zap.String("Result", sayResult))
	return nil
}

// Get Name Activity.
func getNameActivity() (string, error) {
	return "Cadence", nil
}

// Get Greeting Activity.
func getGreetingActivity() (string, error) {
	return "Hello", nil
}

// Say Greeting Activity.
func sayGreetingActivity(greeting string, name string) (string, error) {
	result := fmt.Sprintf("Greeting: %s %s!\n", greeting, name)
	return result, nil
}
