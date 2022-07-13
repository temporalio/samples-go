package zapadapter

import (
	"context"
	"errors"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/workflow"
)

// Workflow is a workflow function which does some logging.
// Important note: workflow logger is replay aware and it won't log during replay.
func Workflow(ctx workflow.Context, name string) error {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("Logging from workflow", "name", name)

	var result interface{}
	err := workflow.ExecuteActivity(ctx, LoggingActivity, name).Get(ctx, &result)
	if err != nil {
		logger.Error("LoggingActivity failed.", "Error", err)
		return err
	}

	err = workflow.ExecuteActivity(ctx, LoggingErrorAcctivity).Get(ctx, &result)
	if err != nil {
		logger.Error("LoggingActivity failed.", "Error", err)
		return err
	}

	logger.Info("Workflow completed.")
	return nil
}

func LoggingActivity(ctx context.Context, name string) error {
	logger := activity.GetLogger(ctx)
	withLogger := logger.(log.WithLogger).With("activity", "LoggingActivity")

	withLogger.Info("Executing LoggingActivity.", "name", name)
	withLogger.Debug("Debugging LoggingActivity.", "value", "important debug data")
	return nil
}

func LoggingErrorAcctivity(ctx context.Context) error {
	logger := activity.GetLogger(ctx)
	logger.Warn("Ignore next error message. It is just for demo purpose.")
	logger.Error("Unable to execute LoggingErrorAcctivity.", "error", errors.New("random error"))
	return nil
}
