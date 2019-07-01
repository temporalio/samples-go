package main

import (
	"context"
	"time"

	"go.uber.org/cadence"
	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

/**
 * This is the cancel activity workflow sample.
 */

// ApplicationName is the task list for this sample
const ApplicationName = "cancelGroup"

// This is registration process where you register all your workflows
// and activity function handlers.
func init() {
	workflow.Register(Workflow)
	activity.Register(activityToBeCacneled)
	activity.Register(cleanupActivity)
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

	err := workflow.ExecuteActivity(ctx, activityToBeCacneled).Get(ctx, nil)
	if err != nil {
		if cadence.IsCanceledError(err) {
			// here, we need to get a new ctx to perform cleanup
			newCtx, _ := workflow.NewDisconnectedContext(ctx)
			err = workflow.ExecuteActivity(newCtx, cleanupActivity).Get(ctx, nil)
			if err != nil {
				logger.Error("Cleanup activity failed", zap.Error(err))
				return err
			}
		} else {
			logger.Error("Activity activityToBeCacneled failed.", zap.Error(err))
			return err
		}
	} else {

	}

	logger.Info("Workflow completed.")

	return nil
}

func activityToBeCacneled(ctx context.Context) error {
	//"operationToCancel---cancel workflow after start. Using command 'cadence --do samples-domain wf cancel -w <WorkflowID>' ", 240
	logger := activity.GetLogger(ctx)
	logger.Info("activity started, you can use ./cancelactivity -m cancel <WorkflowID> or CLI: 'cadence --do samples-domain wf cancel -w <WorkflowID>' to cancel")
	for {
		select {
		case <-time.After(1 * time.Second):
			logger.Info("heartbeating...")
			activity.RecordHeartbeat(ctx, "")
		case <-ctx.Done():
			logger.Info("context is cancelled")
			return nil
		}
	}

	return nil
}

func cleanupActivity(ctx context.Context) error {
	logger := activity.GetLogger(ctx)
	logger.Info("cleanupActivity started")
	return nil
}
