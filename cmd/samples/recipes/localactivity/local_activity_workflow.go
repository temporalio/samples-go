package main

import (
	"context"
	"go.temporal.io/temporal/activity"
	"go.temporal.io/temporal/workflow"
	"go.temporal.io/temporal"
	"go.uber.org/zap"
	"strings"
	"time"
)

/**
 * Sample workflow that uses local activities.
 */

// ApplicationName is the task list for this sample
const ApplicationName = "localActivityGroup"

// SignalName is the signal name that workflow is waiting for
const SignalName = "trigger-signal"

// This is registration process where you register all your workflows
// and activity function handlers.
func init() {
	workflow.Register(ProcessingWorkflow)
	workflow.Register(SignalHandlingWorkflow)

	activity.Register(activityForCondition0)
	activity.Register(activityForCondition1)
	activity.Register(activityForCondition2)
	activity.Register(activityForCondition3)
	activity.Register(activityForCondition4)

	// no need to register local activities
}

type conditionAndAction struct {
	// condition is a function pointer to a local activity
	condition interface{}
	// action is a function pointer to a regular activity
	action interface{}
}

var checks = []conditionAndAction{
	{checkCondition0, activityForCondition0},
	{checkCondition1, activityForCondition1},
	{checkCondition2, activityForCondition2},
	{checkCondition3, activityForCondition3},
	{checkCondition4, activityForCondition4},
}

// ProcessingWorkflow is a workflow that process a given signal data. It evaluates if any conditions are meet for
// the given signal data by using LocalActivity which runs as local function and then schedule activities to handle
// it if the condition is meet. The idea is that you could have many conditions (for example 100 conditions) that needs
// to be evaluated, and only a couple of them will meet the condition and needs to be processed by an activity. Using
// local activity is very efficient in this case because local activity is execute as local function directly by decider
// worker.
func ProcessingWorkflow(ctx workflow.Context, data string) (string, error) {
	logger := workflow.GetLogger(ctx)

	lao := workflow.LocalActivityOptions{
		// use short timeout as local activity is execute as function locally.
		ScheduleToCloseTimeout: time.Second,
	}
	ctx = workflow.WithLocalActivityOptions(ctx, lao)

	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var actionFutures []workflow.Future

	for i, check := range checks {
		var conditionMeet bool
		err := workflow.ExecuteLocalActivity(ctx, check.condition, data).Get(ctx, &conditionMeet)
		if err != nil {
			return "", err
		}

		logger.Sugar().Infof("condition meet for %v: %v", i, conditionMeet)
		if conditionMeet {
			f := workflow.ExecuteActivity(ctx, check.action, data)
			actionFutures = append(actionFutures, f)
		}
	}

	var processResult string
	for _, f := range actionFutures {
		var actionResult string
		if err := f.Get(ctx, &actionResult); err != nil {
			return "", err
		}
		processResult += actionResult
	}

	return processResult, nil
}

// SignalHandlingWorkflow is a workflow that waits on signal and then sends that signal to be processed by a child workflow.
func SignalHandlingWorkflow(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)
	ch := workflow.GetSignalChannel(ctx, SignalName)
	for {
		var signal string
		if more := ch.Receive(ctx, &signal); !more {
			logger.Info("Signal channel closed")
			return temporal.NewCustomError("signal_channel_closed")
		}

		logger.Info("Signal received.", zap.String("signal", signal))

		if signal == "exit" {
			break
		}

		cwo := workflow.ChildWorkflowOptions{
			ExecutionStartToCloseTimeout: time.Minute,
			// TaskStartToCloseTimeout must be larger than all local activity execution time, because DecisionTask won't
			// return until all local activities completed.
			TaskStartToCloseTimeout: time.Second * 30,
		}
		childCtx := workflow.WithChildOptions(ctx, cwo)

		var processResult string
		err := workflow.ExecuteChildWorkflow(childCtx, ProcessingWorkflow, signal).Get(childCtx, &processResult)
		if err != nil {
			return err
		}
		logger.Sugar().Infof("Processed signal: %v, result: %v", signal, processResult)
	}

	return nil
}

func checkCondition0(ctx context.Context, signal string) (bool, error) {
	// some real logic happen here...
	return strings.Contains(signal, "_0_"), nil
}

func checkCondition1(ctx context.Context, signal string) (bool, error) {
	// some real logic happen here...
	return strings.Contains(signal, "_1_"), nil
}

func checkCondition2(ctx context.Context, signal string) (bool, error) {
	// some real logic happen here...
	return strings.Contains(signal, "_2_"), nil
}

func checkCondition3(ctx context.Context, signal string) (bool, error) {
	// some real logic happen here...
	return strings.Contains(signal, "_3_"), nil
}

func checkCondition4(ctx context.Context, signal string) (bool, error) {
	// some real logic happen here...
	return strings.Contains(signal, "_4_"), nil
}

func activityForCondition0(ctx context.Context, signal string) (string, error) {
	activity.GetLogger(ctx).Info("process for condition 0")
	// some real processing logic goes here
	time.Sleep(time.Second * 2)
	return "processed_0", nil
}

func activityForCondition1(ctx context.Context, signal string) (string, error) {
	activity.GetLogger(ctx).Info("process for condition 1")
	// some real processing logic goes here
	time.Sleep(time.Second * 2)
	return "processed_1", nil
}

func activityForCondition2(ctx context.Context, signal string) (string, error) {
	activity.GetLogger(ctx).Info("process for condition 2")
	// some real processing logic goes here
	time.Sleep(time.Second * 2)
	return "processed_2", nil
}

func activityForCondition3(ctx context.Context, signal string) (string, error) {
	activity.GetLogger(ctx).Info("process for condition 3")
	// some real processing logic goes here
	time.Sleep(time.Second * 2)
	return "processed_3", nil
}

func activityForCondition4(ctx context.Context, signal string) (string, error) {
	activity.GetLogger(ctx).Info("process for condition 4")
	// some real processing logic goes here
	time.Sleep(time.Second * 2)
	return "processed_4", nil
}