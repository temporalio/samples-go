package update_cancel

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

const (
	UpdateHandle = "update_handle"
	Done         = "done"
)

func UpdateWorkflow(ctx workflow.Context) error {
	if err := workflow.SetUpdateHandler(
		ctx,
		UpdateHandle,
		func(ctx workflow.Context, sleepTime time.Duration) (time.Duration, error) {
			dt := workflow.Now(ctx)
			workflow.Sleep(ctx, sleepTime)
			return workflow.Now(ctx).Sub(dt), nil
		},
	); err != nil {
		return err
	}

	_ = workflow.GetSignalChannel(ctx, Done).Receive(ctx, nil)
	return ctx.Err()
}
