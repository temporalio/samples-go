package main

import (
	"context"
	"fmt"
	"time"

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
	activity.Register(activityToBeCanceled)
	activity.Register(activityToBeSkipped)
	activity.Register(cleanupActivity)
}

// Workflow workflow decider
func Workflow(ctx workflow.Context) error {
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute * 30,
		HeartbeatTimeout:       time.Second * 5,
		WaitForCancellation:    true,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	logger := workflow.GetLogger(ctx)
	logger.Info("cancel workflow started")

	defer func() {
		newCtx, _ := workflow.NewDisconnectedContext(ctx)
		err := workflow.ExecuteActivity(newCtx, cleanupActivity).Get(ctx, nil)
		if err != nil {
			logger.Error("Cleanup activity failed", zap.Error(err))
		}
	}()

	var result string
	err := workflow.ExecuteActivity(ctx, activityToBeCanceled).Get(ctx, &result)
	logger.Info(fmt.Sprintf("activityToBeCanceled returns %v, %v", result, err))

	_ = workflow.ExecuteActivity(ctx, activityToBeSkipped).Get(ctx, nil)

	logger.Info("Workflow completed.")

	return nil
}

func activityToBeCanceled(ctx context.Context) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("activity started, you can use ./cancelactivity -m cancel -w <WorkflowID> or CLI: 'cadence --do samples-domain wf cancel -w <WorkflowID>' to cancel")
	for {
		select {
		case <-time.After(1 * time.Second):
			logger.Info("heartbeating...")
			activity.RecordHeartbeat(ctx, "")
		case <-ctx.Done():
			logger.Info("context is cancelled")
			return "I am canceled by Done", nil
		}
	}
}

func cleanupActivity(ctx context.Context) error {
	logger := activity.GetLogger(ctx)
	logger.Info("cleanupActivity started")
	return nil
}

func activityToBeSkipped(ctx context.Context) error {
	logger := activity.GetLogger(ctx)
	logger.Info("this activity will be skipped due to cancellation")
	return nil
}
