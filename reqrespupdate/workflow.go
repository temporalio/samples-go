package reqrespupdate

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

const (
	UpdateHandler = "update"
)

var (
	ErrBackoff = errors.New("trying to continue as new")
)

// UppercaseWorkflow is a workflow that accepts requests to uppercase strings
// via updates and provides a response.
func UppercaseWorkflow(ctx workflow.Context, rejectUpdateOnPendingContinueAsNew bool) error {
	// Create and run the uppercaser. We choose to use a separate struct for this
	// to make state management easier.
	u, err := newUppercaser(ctx, rejectUpdateOnPendingContinueAsNew)
	if err == nil {
		err = u.run(ctx)
	}
	return err
}

// UppercaseActivity uppercases the given string.
func UppercaseActivity(ctx context.Context, input string) (string, error) {
	return strings.ToUpper(input), nil
}

// Request is a request to uppercase a string, passed as a update argument to
// UppercaseWorkflow.
type Request struct {
	// String to be uppercased.
	Input string `json:"input"`
}

// Response is a response to a Request. This is returned as a response to the update request
type Response struct {
	Output string `json:"output"`
}

type uppercaser struct {
	workflow.Context
	requestsBeforeContinueAsNew        int
	rejectUpdateOnPendingContinueAsNew bool
}

func newUppercaser(ctx workflow.Context, rejectUpdateOnPendingContinueAsNew bool) (*uppercaser, error) {
	u := &uppercaser{
		// For the main context, we're only going to allow 1 retry and only a 5
		// second schedule-to-close timeout. WARNING: The timeout and retry affect
		// how long this workflow stays open and may prevent it from performing its
		// continue-as-new until timeout occurs and/or retries are finished.
		Context: workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
			ScheduleToCloseTimeout: 5 * time.Second,
			RetryPolicy:            &temporal.RetryPolicy{MaximumAttempts: 2},
		}),
		// We'll allow 500 requests before we continue-as-new the workflow. This is
		// required because the history will grow very large otherwise for an
		// interminable workflow fielding update requests and executing activities.
		requestsBeforeContinueAsNew: 500,
		// If the workflow is trying to continue as new, but update requests are coming
		// in faster than they are handled it is possible the workflow will never be able
		// continue as new. To try to mitigate this the workflow can be set to reject incoming update
		// through the validator. The requester will see this rejection and backoff.
		rejectUpdateOnPendingContinueAsNew: rejectUpdateOnPendingContinueAsNew,
	}
	return u, nil
}

func (u *uppercaser) run(ctx workflow.Context) error {
	var requestCount int
	var pendingUpdates int

	var options workflow.UpdateHandlerOptions
	if u.rejectUpdateOnPendingContinueAsNew {
		options.Validator = func(ctx workflow.Context, request Request) error {
			if requestCount >= u.requestsBeforeContinueAsNew {
				// Rejecting an update in the validator will not persist the update
				// to history which is useful if the history size is growing large.
				return ErrBackoff
			}
			return nil
		}
	}
	// Set update handler
	err := workflow.SetUpdateHandlerWithOptions(ctx, UpdateHandler, func(ctx workflow.Context, request Request) (Response, error) {
		requestCount++
		pendingUpdates++
		defer func() {
			pendingUpdates--
		}()
		var response Response
		err := workflow.ExecuteActivity(u, UppercaseActivity, request.Input).Get(ctx, &response.Output)
		return response, err
	}, options)
	if err != nil {
		return fmt.Errorf("failed setting updatex handler: %w", err)
	}

	// Wait until we can continue as new or are cancelled.
	err = workflow.Await(ctx, func() bool { return requestCount >= u.requestsBeforeContinueAsNew && pendingUpdates == 0 })
	if err != nil {
		return err
	}

	// Continue as new since there were too many responses and there is no pending updates.
	// Note, if update requests come in faster than they
	// are handled, there will not be a moment where the workflow has
	// nothing pending which means this will run forever.
	return workflow.NewContinueAsNewError(u, UppercaseWorkflow, u.rejectUpdateOnPendingContinueAsNew)
}
