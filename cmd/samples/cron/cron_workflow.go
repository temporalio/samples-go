package main

import (
	"context"
	"time"

	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

/**
 * This cron sample workflow will schedule job based on given schedule spec. The schedule spec in this sample demo is
 * very simple, but you could have more complicated scheduler logic that meet your needs.
 */

type (
	// ScheduleSpec specify how the cron job will be scheduled.
	ScheduleSpec struct {
		// How many times you want the cron job to be scheduled.
		JobCount         uint
		ScheduleInterval time.Duration
	}
)

const (
	// ApplicationName is the task list for this sample
	ApplicationName = "cronGroup"

	// timeouts for activity
	scheduleToCloseTimeout = time.Minute * 10
	scheduleToStartTimeout = time.Minute * 10
	startToCloseTimeout    = time.Minute * 10
	heartbeatTimeout       = time.Minute * 10

	// timeout for workflow
	workflowTimeout = time.Minute * 20
	decisionTimeout = time.Minute * 1

	// Every activity execution in workflow increases the size of workflow execution's history. We don't want the history
	// grow to very large because large history is expensive to process. So, in this sample, we will create new workflow
	// for every 10 job runs.
	loopCountBeforeContinueAsNew = 10
)

func (s *ScheduleSpec) getDelayBeforeNextRun() time.Duration {
	// For this sample, we use this naive solution. But you could have your own logic that meets your scheduling requirement.
	return s.ScheduleInterval
}

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
func sampleCronActivity(ctx context.Context, pendingJobCount uint) error {
	activity.GetLogger(ctx).Info("Cron job running.",
		zap.Uint("PendingJobCount", pendingJobCount))
	// ...
	return nil
}

// SampleCronWorkflow workflow decider
func SampleCronWorkflow(ctx workflow.Context, scheduleSpec ScheduleSpec) (err error) {
	if scheduleSpec.JobCount == 0 {
		// should not happen... but if it does, there is nothing to do, since we are done here.
		workflow.GetLogger(ctx).Info("Cron workflow started with 0 JobCount.")
		return nil
	}

	workflow.GetLogger(ctx).Info("Cron workflow started.",
		zap.Duration("IntervalInterval", scheduleSpec.ScheduleInterval),
		zap.Uint("ScheduledCount", scheduleSpec.JobCount))

	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: scheduleToStartTimeout,
		StartToCloseTimeout:    startToCloseTimeout,
		HeartbeatTimeout:       heartbeatTimeout,
	}
	ctx1 := workflow.WithActivityOptions(ctx, ao)

	for i := 0; i < loopCountBeforeContinueAsNew && scheduleSpec.JobCount > 0; i++ {
		scheduleSpec.JobCount--

		sleepDuration := scheduleSpec.getDelayBeforeNextRun()
		workflow.Sleep(ctx, sleepDuration)

		err = workflow.ExecuteActivity(ctx1, sampleCronActivity, scheduleSpec.JobCount).Get(ctx, nil)
		if err != nil {
			// Appropriate retries needed for the workflow business logic.
			// - The activity can be retired on multiple failures look at workflow.ExecuteActivity documentation to
			// see what possible errors it can return.
			// - look at our sample recipes/retryActivity.
			return err
		}
	}

	if scheduleSpec.JobCount == 0 {
		// done with this cron workflow
		workflow.GetLogger(ctx).Info("Cron workflow completed.")
		return nil
	}

	// schedule next cron job
	ctx = workflow.WithExecutionStartToCloseTimeout(ctx, workflowTimeout)
	ctx = workflow.WithWorkflowTaskStartToCloseTimeout(ctx, decisionTimeout)

	return workflow.NewContinueAsNewError(ctx, SampleCronWorkflow, scheduleSpec)
}
