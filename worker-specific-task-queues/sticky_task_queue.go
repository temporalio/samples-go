package activities_sticky_queues

import (
	"context"
)

type StickyTaskQueue struct {
	TaskQueue string
}

// GetStickyTaskQueue is an activity to get a hosts unique task queue.
func (q StickyTaskQueue) GetStickyTaskQueue(ctx context.Context) (string, error) {
	return q.TaskQueue, nil
}
