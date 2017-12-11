package main

import (
	"fmt"
	"time"

	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

/**
 * This greetings sample workflow executes 3 activities in sequential. It gets greeting and name from 2 different activities,
 * and then pass greeting and name as input to a 3rd activity to generate final greetings.
 */

// ApplicationName is the task list for this sample
const ApplicationName = "greetingsGroup"

// This is registration process where you register all your workflows
// and activity function handlers.
func init() {
	workflow.Register(SampleGreetingsWorkflow)
	activity.Register(getGreetingActivity)
	activity.Register(getNameActivity)
	activity.Register(sayGreetingActivity)
}

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
	err := workflow.ExecuteActivity(ctx, getGreetingActivity).Get(ctx, &greetResult)
	if err != nil {
		logger.Error("Get greeting failed.", zap.Error(err))
		return err
	}

	// Get Name.
	var nameResult string
	err = workflow.ExecuteActivity(ctx, getNameActivity).Get(ctx, &nameResult)
	if err != nil {
		logger.Error("Get name failed.", zap.Error(err))
		return err
	}

	// Say Greeting.
	var sayResult string
	err = workflow.ExecuteActivity(ctx, sayGreetingActivity, greetResult, nameResult).Get(ctx, &sayResult)
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
