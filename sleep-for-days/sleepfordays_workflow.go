package helloworld

import (
	"context"
	"fmt"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/workflow"
)

const CompleteSignal = "complete"

func SleepForDaysWorkflow(ctx workflow.Context, days int) (string, error) {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	})

	isComplete := false
	sigChan := workflow.GetSignalChannel(ctx, CompleteSignal)
	workflow.Go(ctx, func(ctx workflow.Context) {
		sigChan.Receive(ctx, &isComplete)
	})

	for !isComplete {
		workflow.ExecuteActivity(ctx, SendEmailActivity, fmt.Sprintf("Sleeping for %d days", days)).IsReady()
		workflow.Sleep(ctx, time.Hour*24*time.Duration(days))
	}

	return "done", nil
}

// A stub Activity for sending an email.
func SendEmailActivity(ctx context.Context, msg string) error {
	activity.GetLogger(ctx).Info(fmt.Sprintf(`Sending email: "%v"\n`, msg))
	return nil
}
