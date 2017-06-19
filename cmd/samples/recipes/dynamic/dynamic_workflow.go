package main

import (
	"fmt"
	"time"

	"go.uber.org/cadence"
	"go.uber.org/zap"
)

/**
 * The purpose of this sample is to demonstrate invocation of workflows and activities using name rather than strongly
 * typed function.
 */

// ApplicationName is the task list for this sample
const ApplicationName = "dynamicGroup"

// GreetingsWorkflowName name used when workflow function is registered during init.  We use the fully qualified name to function
const GreetingsWorkflowName = "main.SampleGreetingsWorkflow"

// Activity names used when activity function is registered during init.  We use the fully qualified name to function
const getNameActivityName = "main.getNameActivity"
const getGreetingActivityName = "main.getGreetingActivity"
const sayGreetingActivityName = "main.sayGreetingActivity"

// This is registration process where you register all your workflows
// and activity function handlers.
func init() {
	cadence.RegisterWorkflow(SampleGreetingsWorkflow)
	cadence.RegisterActivity(getGreetingActivity)
	cadence.RegisterActivity(getNameActivity)
	cadence.RegisterActivity(sayGreetingActivity)
}

// SampleGreetingsWorkflow Workflow Decider.
func SampleGreetingsWorkflow(ctx cadence.Context) error {
	// Get Greeting.
	ao := cadence.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
		HeartbeatTimeout:       time.Second * 20,
	}
	ctx = cadence.WithActivityOptions(ctx, ao)

	logger := cadence.GetLogger(ctx)
	var greetResult string
	err := cadence.ExecuteActivity(ctx, getGreetingActivityName).Get(ctx, &greetResult)
	if err != nil {
		logger.Error("Get greeting failed.", zap.Error(err))
		return err
	}

	// Get Name.
	var nameResult string
	err = cadence.ExecuteActivity(ctx, getNameActivityName).Get(ctx, &nameResult)
	if err != nil {
		logger.Error("Get name failed.", zap.Error(err))
		return err
	}

	// Say Greeting.
	var sayResult string
	err = cadence.ExecuteActivity(ctx, sayGreetingActivityName, greetResult, nameResult).Get(ctx, &sayResult)
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
