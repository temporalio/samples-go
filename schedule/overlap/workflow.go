package overlap

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

const (
	TaskQueue   = "schedule-overlap-policy"
	RunDuration = 5 * time.Second
)

// PolicyWorkflow stays open longer than the Schedule interval so every Action
// after the first one must be handled by the configured Overlap Policy.
func PolicyWorkflow(ctx workflow.Context, policy string) (string, error) {
	info := workflow.GetInfo(ctx)
	logger := workflow.GetLogger(ctx)
	startedAt := workflow.Now(ctx)
	scheduledStartTimeKey := temporal.NewSearchAttributeKeyTime("TemporalScheduledStartTime")
	scheduledAt, _ := workflow.GetTypedSearchAttributes(ctx).GetTime(scheduledStartTimeKey)
	logger.Info("START",
		"Policy", policy,
		"WorkflowID", info.WorkflowExecution.ID,
		"RunID", info.WorkflowExecution.RunID,
		"TemporalScheduledStartTime", scheduledAt,
		"StartedAt", startedAt,
		"ScheduleToStartDelay", startedAt.Sub(scheduledAt),
	)

	if err := workflow.Sleep(ctx, RunDuration); err != nil {
		logger.Info("CANCELED",
			"Policy", policy,
			"WorkflowID", info.WorkflowExecution.ID,
			"RunID", info.WorkflowExecution.RunID,
		)
		return "", err
	}

	result := fmt.Sprintf("%s run %s completed", policy, info.WorkflowExecution.RunID)
	logger.Info("COMPLETE",
		"Policy", policy,
		"WorkflowID", info.WorkflowExecution.ID,
		"RunID", info.WorkflowExecution.RunID,
	)
	return result, nil
}
