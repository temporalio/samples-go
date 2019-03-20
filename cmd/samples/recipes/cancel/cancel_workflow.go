package main

import (
	"context"
	"fmt"
	"time"

	"strconv"

	"go.uber.org/cadence"
	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

/**
 * This is the hello world workflow sample.
 */

// ApplicationName is the task list for this sample
const ApplicationName = "cancelGroup"

// This is registration process where you register all your workflows
// and activity function handlers.
func init() {
	workflow.Register(Workflow)
	activity.Register(cancelActivity)
}

// Workflow workflow decider
func Workflow(ctx workflow.Context) error {
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute * 30,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("cancel workflow started")
	err := workflow.ExecuteActivity(ctx, cancelActivity, "dirtyOperation", 1).Get(ctx, nil)
	if err != nil {
		logger.Error("Activity failed, not expected", zap.Error(err))
		return err
	}

	err = workflow.ExecuteActivity(ctx, cancelActivity, "operationToCancel---cancel workflow after start. Using command 'cadence --do samples-domain wf cancel -w <WorkflowID>' ", 240).Get(ctx, nil)
	if err != nil {
		if cadence.IsCanceledError(err) {
			// here, we need to get a new ctx to perform cleanup
			newCtx, _ := workflow.NewDisconnectedContext(ctx)
			// MyCancelActivity will delete that record from DB
			err = workflow.ExecuteActivity(newCtx, cancelActivity, "cleanUp", 1).Get(ctx, nil)
			if err != nil {
				logger.Error("Cleanup activity failed, not expected", zap.Error(err))
				return err
			}
		} else {
			logger.Error("Activity failed.", zap.Error(err))
			return err
		}
	}

	//try to use old ctx, it won't be executed. Instead, it will return an canceled error
	err = workflow.ExecuteActivity(ctx, cancelActivity, "tryOldContext", 1).Get(ctx, nil)
	if err != nil {
		logger.Error("Activity failed, this is expected", zap.Error(err))
		// if returning this err here, the workflow will finished as Canceled.
		//return err
	}

	logger.Info("Workflow completed.")

	return nil
}

func cancelActivity(ctx context.Context, name string, waitSecs int) error {
	fmt.Println("first line in cancelActivity " + name)
	logger := activity.GetLogger(ctx)
	logger.Info("activity started")
	select {
	case <-time.After(time.Duration(waitSecs) * time.Second):
		logger.Info("wake up after sleep for " + strconv.Itoa(waitSecs) + " seconds")
	case <-ctx.Done():
		logger.Info("context is cancelled")
	}
	return nil
}
