package nondeterminism

import (
	"context"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/workflow"
)

/*
For all of the samples below, "nde" stands for non-deterministic error.
*/

/*
The following similar scenarios DO NOT cause an NDE:
- changing the activity's parameter name
- changing an activity past the last checkpoint of the workflow, since the
replay will not execute beyond the current checkpoint
*/

// WorkflowChangingActivityName shows how you can run into an nde if you just
// change the name of an existing activity (e.g. if the previous name was not
// representative). It is meant to show that refactoring must be handled with
// care.
//
// Once the workflow reaches the sleep, shut down the worker, replace the
// activity name with the new one and re-run the worker. To recover from this
// nde, revert the activity name back to its original value and restart the
// worker.
func WorkflowChangingActivityName(ctx workflow.Context, parameter string) (string, error) {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("HelloWorld workflow started", "parameter", parameter)

	// replace ActivityOriginalName with ActivityNewName during the sleep to
	// get an NDE.
	workflow.ExecuteActivity(ctx, ActivityOriginalName, parameter).Get(ctx, nil)
	// workflow.ExecuteActivity(ctx, ActivityNewName, parameter).Get(ctx, nil)
	logger.Info("First activity completed")

	logger.Info("sleeping...")
	workflow.Sleep(ctx, 10*time.Second)

	// replacing this during the sleep will not cause NDE
	// workflow.ExecuteActivity(ctx, ActivityOriginalName, parameter).Get(ctx, nil)
	workflow.ExecuteActivity(ctx, ActivityNewName, parameter).Get(ctx, nil)
	logger.Info("Second activitys completed")

	return "done", nil
}

func ActivityOriginalName(ctx context.Context, parameter string) error {
	logger := activity.GetLogger(ctx)
	logger.Info("Activity hello", "parameter", parameter)
	return nil
}

// ActivityNewName is identical to ActivityOriginal in terms of signature and
// implementation.
func ActivityNewName(ctx context.Context, parameter string) error {
	logger := activity.GetLogger(ctx)
	logger.Info("Activity bye", "parameter", parameter)
	return nil
}
