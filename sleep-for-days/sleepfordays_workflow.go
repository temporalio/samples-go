package sleepfordays

import (
	"context"
	"fmt"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/workflow"
)

func SleepForDaysWorkflow(ctx workflow.Context) (string, error) {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	})

	isComplete := false
	sigChan := workflow.GetSignalChannel(ctx, "complete")

	for !isComplete {
		workflow.ExecuteActivity(ctx, SendEmailActivity, "Sleeping for 30 days")
		selector := workflow.NewSelector(ctx)
		selector.AddFuture(workflow.NewTimer(ctx, time.Hour*24*30), func(f workflow.Future) {})
		selector.AddReceive(sigChan, func(c workflow.ReceiveChannel, more bool) {
			isComplete = true
		})
		selector.Select(ctx)
	}

	return "done", nil
}

// A stub Activity for sending an email.
func SendEmailActivity(ctx context.Context, msg string) error {
	activity.GetLogger(ctx).Info(fmt.Sprintf(`Sending email: "%v"\n`, msg))
	return nil
}
