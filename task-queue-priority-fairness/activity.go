package task_queue_priority_fairness

import (
	"context"
	"time"

	"go.temporal.io/sdk/activity"
)

func ProcessRenderJob(ctx context.Context, job RenderJob) (RenderResult, error) {
	startedAt := time.Now().UTC()

	logger := activity.GetLogger(ctx)
	logger.Info(
		"Started render job",
		"started_at", startedAt,
		"priority", job.PriorityKey,
		"tenant", job.Tenant,
		"weight", job.FairnessWeight,
		"kind", job.Kind,
		"job_id", job.JobID,
	)

	time.Sleep(150 * time.Millisecond)

	return RenderResult{
		StartedAt:      startedAt,
		JobID:          job.JobID,
		Tenant:         job.Tenant,
		Kind:           job.Kind,
		PriorityKey:    job.PriorityKey,
		FairnessKey:    job.FairnessKey,
		FairnessWeight: job.FairnessWeight,
	}, nil
}
