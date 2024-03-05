package async_update

import (
	"context"
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"
)

const (
	ProcessUpdateName = "process"
	Done              = "done"
)

func ProcessWorkflow(ctx workflow.Context) (int, error) {
	logger := workflow.GetLogger(ctx)
	// inProgressJobs is used to keep track of the number of jobs currently being processed.
	inProgressJobs := 0
	// processedJobs is used to keep track of the number of jobs processed so far.
	processedJobs := 0
	// closing is used to keep track of whether the workflow is closing.
	closing := false

	if err := workflow.SetUpdateHandlerWithOptions(
		ctx,
		ProcessUpdateName,
		func(ctx workflow.Context, s string) (string, error) {
			inProgressJobs++
			processedJobs++
			defer func() {
				inProgressJobs--
			}()
			logger.Debug("Processing job", "job", s)
			ao := workflow.ActivityOptions{
				StartToCloseTimeout: 10 * time.Second,
			}
			ctx = workflow.WithActivityOptions(ctx, ao)
			var result string
			err := workflow.ExecuteActivity(ctx, Activity, s).Get(ctx, &result)
			logger.Debug("Processed job", "job", s)
			return result, err
		},
		workflow.UpdateHandlerOptions{
			Validator: func(s string) error {
				logger.Debug("Validating job", "job", s, "inProgressJobs", inProgressJobs)
				if inProgressJobs >= 5 {
					return fmt.Errorf("too many in progress jobs: %d", inProgressJobs)
				} else if closing {
					return fmt.Errorf("workflow is closing")
				}
				return nil
			},
		},
	); err != nil {
		return 0, err
	}

	_ = workflow.GetSignalChannel(ctx, Done).Receive(ctx, nil)
	logger.Debug("Closing workflow, draining in progress jobs", "inProgressJobs", inProgressJobs)
	// set closing to true to indicate that the workflow is closing.
	// no more new jobs are allowed, but the existing jobs will be processed.
	closing = true
	workflow.Await(ctx, func() bool {
		return inProgressJobs == 0
	})
	logger.Debug("All jobs processed, workflow can now close")
	return processedJobs, ctx.Err()
}

func Activity(ctx context.Context, name string) (string, error) {
	// Simulate a long running activity
	time.Sleep(5 * time.Second)
	return "Hello " + name + "!", nil
}
