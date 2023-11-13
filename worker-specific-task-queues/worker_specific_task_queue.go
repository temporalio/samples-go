package worker_specific_task_queues

import (
	"context"
)

type WorkerSpecificTaskQueue struct {
	TaskQueue string
}

// GetWorkerSpecificTaskQueue is an activity to get a hosts unique task queue.
func (q WorkerSpecificTaskQueue) GetWorkerSpecificTaskQueue(ctx context.Context) (string, error) {
	return q.TaskQueue, nil
}
