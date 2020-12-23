package metrics

import (
	"context"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/workflow"
)

func Workflow(ctx workflow.Context) error {
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("Metrics workflow started.")

	scheduledTimeNanos := workflow.Now(ctx).UnixNano()
	_ = workflow.Sleep(ctx, 500*time.Millisecond)
	err := workflow.ExecuteActivity(ctx, Activity, scheduledTimeNanos).Get(ctx, nil)
	if err != nil {
		logger.Error("Activity failed.", "Error", err)
		return err
	}

	logger.Info("Metrics workflow completed.")
	return nil
}

func Activity(ctx context.Context, scheduledTimeNanos int64) error {
	logger := activity.GetLogger(ctx)

	var err error
	metricsScope := activity.GetMetricsScope(ctx)
	metricsScope, sw := recordActivityStart(metricsScope, "metrics.Activity", scheduledTimeNanos)
	defer func() { recordActivityEnd(metricsScope, sw, err) }()

	time.Sleep(time.Second)
	logger.Info("Metrics reported.")
	return err
}
