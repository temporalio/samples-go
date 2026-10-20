package batch_queue

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"go.temporal.io/sdk/workflow"
)

const (
	SIGNAL_READ_VALS    = "read_vals"
	SIGNAL_COMMIT_BATCH = "commit_batch"
)

func GetAccumulateAndBatchWorkflowID() string {
	return "AccumulateAndBatchWorkflowID"
}

type WorkflowTicker struct {
	d time.Duration
	C workflow.Channel
}

func NewWorkflowTicker(ctx workflow.Context, d time.Duration) *WorkflowTicker {
	wt := WorkflowTicker{
		d: d,
		C: workflow.NewChannel(ctx),
	}

	workflow.Go(ctx, func(ctx workflow.Context) {
		for {
			t := workflow.NewTimer(ctx, d)
			err := t.Get(ctx, nil)
			if err != nil {
				workflow.GetLogger(ctx).Error("timer failed, restarting...", "error", err)
				continue
			}
			wt.C.Send(ctx, nil)
		}
	})

	return &wt
}

// AccumulateAndBatchWorkflow keeps accumulating data via signals. A timer
// decides when the accumulated data should be processed as a batch.
func AccumulateAndBatchWorkflow(ctx workflow.Context, vals []string) error {

	logger := workflow.GetLogger(ctx)
	logger.Info("AccumulateAndBatchWorkflow workflow started", "vals", vals)

	// listen for incoming data
	readValsCh := workflow.GetSignalChannel(ctx, SIGNAL_READ_VALS)

	// batch every 10 seconds
	t := NewWorkflowTicker(ctx, 10*time.Second)

	eventsReceived := 0

	selector := workflow.NewSelector(ctx)
	selector.AddReceive(readValsCh, func(c workflow.ReceiveChannel, more bool) {
		eventsReceived += 1
		var incoming string
		c.Receive(ctx, &incoming)
		vals = append(vals, incoming)
	})
	selector.AddReceive(t.C, func(c workflow.ReceiveChannel, more bool) {
		eventsReceived += 1
		c.Receive(ctx, nil)
		logger.Info("commiting batch...")

		ao := workflow.ActivityOptions{
			StartToCloseTimeout: 1 * time.Minute,
		}
		actx := workflow.WithActivityOptions(ctx, ao)
		err := workflow.ExecuteActivity(actx, WriteBatchToFile, vals).Get(ctx, nil)
		if err != nil {
			// Couldn't write the batch, so do not discard the data. Note that
			// activity should be idempotent, so you may need to dedupe. This
			// example doesn't dedupe.
			logger.Error("failed to write batch", "error", err)
			return
		}

		// Discard the written data.
		vals = []string{}
	})

	// Batch when you received enough events. Using a low number just to
	// illustrate the batching, you can go much higher.
	for {
		selector.Select(ctx)
		if eventsReceived > 15 {
			// Pass in the existing vals since the signal history is no longer
			// available.
			return workflow.NewContinueAsNewError(ctx, AccumulateAndBatchWorkflow, vals)
		}
	}
}

func WriteBatchToFile(ctx context.Context, vals []string) error {
	// Write the values to this file. We can compare values here to values sent
	// to the workflow to see the durability.
	f, err := os.OpenFile("values_received.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	f.WriteString(strings.Join(vals, "\n"))
	f.WriteString("\n")

	return nil
}
