package streams

import (
	"context"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/contrib/workflowstreams"
)

// ChargeCard (scenario 1) charges a card and publishes fine-grained progress
// events back to its parent workflow's stream. It uses NewClientFromActivity,
// which targets the workflow that scheduled the activity.
func ChargeCard(ctx context.Context, orderID string) (string, error) {
	c, err := workflowstreams.NewClientFromActivity(ctx, workflowstreams.Options{
		BatchInterval: 200 * time.Millisecond,
	})
	if err != nil {
		return "", err
	}
	// Close flushes any buffered items before the activity returns.
	defer func() { _ = c.Close(ctx) }()

	progress := c.Topic(TopicProgress)
	progress.Publish(ProgressEvent{Message: "charging card..."}, false)

	activity.GetLogger(ctx).Info("charging card", "orderID", orderID)
	time.Sleep(time.Second)

	progress.Publish(ProgressEvent{Message: "card charged"}, false)
	return "charge-" + orderID, nil
}
