// @@@SNIPSTART samples-go-schedule-workflow
package schedule

import (
	"context"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/workflow"
)

// SampleScheduleWorkflow executes on the given schedule
func SampleScheduleWorkflow(ctx workflow.Context) error {

	workflow.GetLogger(ctx).Info("Schedule workflow started.", "StartTime", workflow.Now(ctx))

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx1 := workflow.WithActivityOptions(ctx, ao)

	info := workflow.GetInfo(ctx1)

	// Workflow Executions started by a Schedule have the following additional properties appended to their search attributes
	scheduledByIDPayload := info.SearchAttributes.IndexedFields["TemporalScheduledById"]
	var scheduledByID string
	err := converter.GetDefaultDataConverter().FromPayload(scheduledByIDPayload, &scheduledByID)
	if err != nil {
		return err
	}

	startTimePayload := info.SearchAttributes.IndexedFields["TemporalScheduledStartTime"]
	var startTime time.Time
	err = converter.GetDefaultDataConverter().FromPayload(startTimePayload, &startTime)
	if err != nil {
		return err
	}

	err = workflow.ExecuteActivity(ctx1, DoSomething, scheduledByID, startTime).Get(ctx, nil)
	if err != nil {
		workflow.GetLogger(ctx).Error("schedule workflow failed.", "Error", err)
		return err
	}

	return nil
}

// DoSomething is an Activity
func DoSomething(ctx context.Context, scheduleByID string, startTime time.Time) error {
	activity.GetLogger(ctx).Info("Schedulde job running.", "scheduleByID", scheduleByID, "startTime", startTime)
	// Query database, call external API, or do any other non-deterministic action.
	return nil
}

// @@@SNIPEND
