package streams

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/contrib/workflowstreams"
	"go.temporal.io/sdk/workflow"
)

// drainDelay gives subscribers a moment to poll for the final published items
// before the workflow completes and the stream stops serving polls.
const drainDelay = 500 * time.Millisecond

// OrderWorkflow (scenario 1) publishes status events directly from workflow code
// and runs an activity that publishes fine-grained progress events to the same
// stream. A subscriber consumes both topics.
func OrderWorkflow(ctx workflow.Context, input OrderInput) (string, error) {
	// NewWorkflowStream is workflow-safe; workflowcheck only flags it because it copies
	// maps when restoring continue-as-new state, which is order-independent.
	//workflowcheck:ignore
	stream, err := workflowstreams.NewWorkflowStream(ctx, input.StreamState)
	if err != nil {
		return "", err
	}
	status := stream.Topic(TopicStatus)
	progress := stream.Topic(TopicProgress)

	if err := status.Publish(StatusEvent{Kind: "received", OrderID: input.OrderID}); err != nil {
		return "", err
	}

	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
	})
	var chargeID string
	if err := workflow.ExecuteActivity(ctx, ChargeCard, input.OrderID).Get(ctx, &chargeID); err != nil {
		return "", err
	}

	if err := status.Publish(StatusEvent{Kind: "shipped", OrderID: input.OrderID}); err != nil {
		return "", err
	}
	if err := progress.Publish(ProgressEvent{Message: "charge id: " + chargeID}); err != nil {
		return "", err
	}
	if err := status.Publish(StatusEvent{Kind: "complete", OrderID: input.OrderID}); err != nil {
		return "", err
	}

	_ = workflow.Sleep(ctx, drainDelay)
	return chargeID, nil
}

// PipelineWorkflow (scenario 2) publishes a sequence of stage events with delays
// between them, giving a subscriber time to disconnect and reconnect mid-stream.
func PipelineWorkflow(ctx workflow.Context, input PipelineInput) (string, error) {
	//workflowcheck:ignore order-independent map copy in NewWorkflowStream; see OrderWorkflow
	stream, err := workflowstreams.NewWorkflowStream(ctx, input.StreamState)
	if err != nil {
		return "", err
	}
	status := stream.Topic(TopicStatus)

	stages := []string{"validating", "loading data", "transforming", "writing output", "verifying", "complete"}
	for _, stage := range stages {
		if err := status.Publish(StageEvent{Stage: stage}); err != nil {
			return "", err
		}
		if stage != "complete" {
			_ = workflow.Sleep(ctx, 2*time.Second)
		}
	}

	_ = workflow.Sleep(ctx, drainDelay)
	return "pipeline " + input.PipelineID + " done", nil
}

// HubWorkflow (scenario 3) does no work of its own; it exists only to host the
// stream for an external publisher and shuts down on a close signal.
func HubWorkflow(ctx workflow.Context, input HubInput) (string, error) {
	//workflowcheck:ignore order-independent map copy in NewWorkflowStream; see OrderWorkflow
	if _, err := workflowstreams.NewWorkflowStream(ctx, input.StreamState); err != nil {
		return "", err
	}

	closed := false
	workflow.Go(ctx, func(ctx workflow.Context) {
		workflow.GetSignalChannel(ctx, CloseSignal).Receive(ctx, nil)
		closed = true
	})
	if err := workflow.Await(ctx, func() bool { return closed }); err != nil {
		return "", err
	}

	_ = workflow.Sleep(ctx, drainDelay)
	return "hub " + input.HubID + " closed", nil
}

// TickerWorkflow (scenario 4) publishes a long run of tick events and bounds the
// log by periodically truncating everything but the most recent keepLast items.
// Fast subscribers see every tick; subscribers that fall behind the truncation
// point silently jump forward to the new base offset.
func TickerWorkflow(ctx workflow.Context, input TickerInput) (string, error) {
	count := input.Count
	if count == 0 {
		count = 50
	}
	keepLast := input.KeepLast
	if keepLast == 0 {
		keepLast = 10
	}
	truncateEvery := input.TruncateEvery
	if truncateEvery == 0 {
		truncateEvery = 5
	}
	interval := input.Interval
	if interval == 0 {
		interval = 200 * time.Millisecond
	}

	//workflowcheck:ignore order-independent map copy in NewWorkflowStream; see OrderWorkflow
	stream, err := workflowstreams.NewWorkflowStream(ctx, input.StreamState)
	if err != nil {
		return "", err
	}
	tick := stream.Topic(TopicTick)

	published := 0
	for n := 0; n < count; n++ {
		if err := tick.Publish(TickEvent{N: n}); err != nil {
			return "", err
		}
		published++
		_ = workflow.Sleep(ctx, interval)

		if published%truncateEvery == 0 && published > keepLast {
			if err := stream.Truncate(int64(published - keepLast)); err != nil {
				return "", err
			}
		}
	}

	_ = workflow.Sleep(ctx, drainDelay)
	return fmt.Sprintf("ticker emitted %d events", published), nil
}

// LLMWorkflow (scenario 5) hosts the stream while a streaming activity owns the
// non-deterministic OpenAI call and publishes token deltas back to subscribers.
func LLMWorkflow(ctx workflow.Context, input LLMInput) (string, error) {
	//workflowcheck:ignore order-independent map copy in NewWorkflowStream; see OrderWorkflow
	if _, err := workflowstreams.NewWorkflowStream(ctx, input.StreamState); err != nil {
		return "", err
	}

	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 2 * time.Minute,
	})
	var result string
	if err := workflow.ExecuteActivity(ctx, StreamCompletion, input).Get(ctx, &result); err != nil {
		return "", err
	}

	_ = workflow.Sleep(ctx, drainDelay)
	return result, nil
}
