package task_queue_priority_fairness

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

const WorkflowTaskQueue = "task-queue-priority-fairness-workflow"
const ActivityTaskQueue = "task-queue-priority-fairness-activity"

const (
	TenantLargeStudio  = "large-studio"
	TenantSmallStudioA = "small-studio-a"
	TenantSmallStudioB = "small-studio-b"
	TenantPremiumMedia = "premium-media"
)

var TenantWeights = map[string]float64{
	TenantLargeStudio:  1.0,
	TenantSmallStudioA: 1.0,
	TenantSmallStudioB: 1.0,
	TenantPremiumMedia: 3.0,
}

type RenderJob struct {
	JobID          string
	Tenant         string
	Kind           string
	PriorityKey    int
	FairnessKey    string
	FairnessWeight float64
}

type RenderResult struct {
	StartedAt      time.Time
	JobID          string
	Tenant         string
	Kind           string
	PriorityKey    int
	FairnessKey    string
	FairnessWeight float64
}

type Summary struct {
	PriorityObserved         bool
	FairnessObserved         bool
	WeightedFairnessObserved bool
	FirstNormalStartedAt     time.Time
	LastUrgentStartedAt      time.Time
}

func RenderWorkflow(ctx workflow.Context, jobs []RenderJob) ([]RenderResult, error) {
	futures := make([]workflow.Future, 0, len(jobs))

	for _, job := range jobs {
		ao := workflow.ActivityOptions{
			TaskQueue:           ActivityTaskQueue,
			StartToCloseTimeout: time.Minute,
			Priority: temporal.Priority{
				PriorityKey:    job.PriorityKey,
				FairnessKey:    job.FairnessKey,
				FairnessWeight: float32(job.FairnessWeight),
			},
		}
		activityCtx := workflow.WithActivityOptions(ctx, ao)
		future := workflow.ExecuteActivity(activityCtx, ProcessRenderJob, job)
		futures = append(futures, future)
	}

	results := make([]RenderResult, 0, len(futures))
	for _, future := range futures {
		var result RenderResult
		if err := future.Get(ctx, &result); err != nil {
			return nil, err
		}
		results = append(results, result)
	}

	return results, nil
}

func BuildJobs() []RenderJob {
	jobs := make([]RenderJob, 0, 41)
	jobs = append(jobs, makeJobs(TenantLargeStudio, 18, "normal-render", 3)...)
	jobs = append(jobs, makeJobs(TenantSmallStudioA, 3, "normal-render", 3)...)
	jobs = append(jobs, makeJobs(TenantSmallStudioB, 3, "normal-render", 3)...)
	jobs = append(jobs, makeJobs(TenantPremiumMedia, 9, "normal-render", 3)...)
	jobs = append(jobs, makeJobs(TenantLargeStudio, 4, "background-archive", 5)...)
	jobs = append(jobs, makeJobs(TenantPremiumMedia, 2, "urgent-preview", 1)...)
	jobs = append(jobs, makeJobs(TenantSmallStudioA, 2, "urgent-preview", 1)...)
	return jobs
}

func makeJobs(tenant string, count int, kind string, priorityKey int) []RenderJob {
	jobs := make([]RenderJob, 0, count)
	for i := 0; i < count; i++ {
		jobs = append(jobs, RenderJob{
			JobID:          fmt.Sprintf("%s-%s-%02d", tenant, kind, i),
			Tenant:         tenant,
			Kind:           kind,
			PriorityKey:    priorityKey,
			FairnessKey:    tenant,
			FairnessWeight: TenantWeights[tenant],
		})
	}
	return jobs
}
