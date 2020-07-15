package cron

import (
	"context"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

/**
 * This cron sample workflow will schedule job based on given schedule spec. The schedule spec in this sample demo is
 * very simple, but you could have more complicated scheduler logic that meet your needs.
 */

// Cron sample job activity.
func SampleCronActivity(ctx context.Context, beginTime, endTime time.Time) error {
	activity.GetLogger(ctx).Info("Cron job running.", zap.Time("beginTime_exclude", beginTime), zap.Time("endTime_include", endTime))
	// ...
	return nil
}

// SampleCronResult used to return data from one cron run to next cron run.
type SampleCronResult struct {
	EndTime time.Time
}

// SampleCronWorkflow is the sample cron workflow.
func SampleCronWorkflow(ctx workflow.Context) (*SampleCronResult, error) {
	workflow.GetLogger(ctx).Info("Cron workflow started.", zap.Time("StartTime", workflow.Now(ctx)))

	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
	}
	ctx1 := workflow.WithActivityOptions(ctx, ao)

	startTime := time.Time{} // start from 0 time for first cron job
	if workflow.HasLastCompletionResult(ctx) {
		var lastResult SampleCronResult
		if err := workflow.GetLastCompletionResult(ctx, &lastResult); err == nil {
			startTime = lastResult.EndTime
		}
	}

	endTime := workflow.Now(ctx)

	err := workflow.ExecuteActivity(ctx1, SampleCronActivity, startTime, endTime).Get(ctx, nil)

	if err != nil {
		// cron job failed. but next cron should continue to be scheduled by server
		workflow.GetLogger(ctx).Error("Cron job failed.", zap.Error(err))
		return nil, err
	}

	return &SampleCronResult{EndTime: endTime}, nil
}
