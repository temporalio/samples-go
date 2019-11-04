package main

import (
	"context"
	"time"

	"go.temporal.io/temporal/activity"
	"go.temporal.io/temporal/workflow"
	"go.uber.org/zap"
)

/**
 * This cron sample workflow will schedule job based on given schedule spec. The schedule spec in this sample demo is
 * very simple, but you could have more complicated scheduler logic that meet your needs.
 */

const (
	// timeout for activity task from put in queue to started
	activityScheduleToStartTimeout = time.Second * 10
	// timeout for activity from start to complete
	activityStartToCloseTimeout = time.Minute

	// WorkflowStartToCloseTimeout (from workflow start to workflow close)
	WorkflowStartToCloseTimeout = time.Minute * 20
	// DecisionTaskStartToCloseTimeout (from decision task started to decision task completed, usually very short)
	DecisionTaskStartToCloseTimeout = time.Second * 10
)

//
// This is registration process where you register all your workflows
// and activity function handlers.
//
func init() {
	workflow.Register(SampleCronWorkflow)
	activity.Register(sampleCronActivity)
}

//
// Cron sample job activity.
//
func sampleCronActivity(ctx context.Context, beginTime, endTime time.Time) error {
	activity.GetLogger(ctx).Info("Cron job running.", zap.Time("beginTime_exclude", beginTime), zap.Time("endTime_include", endTime))
	// ...
	return nil
}

// SampleCronResult used to return data from one cron run to next cron run.
type SampleCronResult struct {
	EndTime time.Time
}

// SampleCronWorkflow workflow decider
func SampleCronWorkflow(ctx workflow.Context) (*SampleCronResult, error) {
	workflow.GetLogger(ctx).Info("Cron workflow started.", zap.Time("StartTime", workflow.Now(ctx)))

	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: activityScheduleToStartTimeout,
		StartToCloseTimeout:    activityStartToCloseTimeout,
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

	err := workflow.ExecuteActivity(ctx1, sampleCronActivity, startTime, endTime).Get(ctx, nil)

	if err != nil {
		// cron job failed. but next cron should continue to be scheduled by Cadence server
		workflow.GetLogger(ctx).Error("Cron job failed.", zap.Error(err))
		return nil, err
	}

	return &SampleCronResult{EndTime: endTime}, nil
}
