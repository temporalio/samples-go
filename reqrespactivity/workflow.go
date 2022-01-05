package reqrespactivity

import (
	"context"
	"strings"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// UppercaseWorkflow is a workflow that accepts requests to uppercase strings
// via signals and provides responses via a callback response activity.
//
// The "request" signal accepts a Request.
func UppercaseWorkflow(ctx workflow.Context) error {
	// Create and run the uppercaser. We choose to use a separate struct for this
	// to make state management easier.
	u, err := newUppercaser(ctx)
	if err == nil {
		err = u.run()
	}
	return err
}

// UppercaseActivity uppercases the given string.
func UppercaseActivity(ctx context.Context, input string) (string, error) {
	return strings.ToUpper(input), nil
}

// Request is a request to uppercase a string, passed as a signal argument to
// UppercaseWorkflow.
type Request struct {
	// ID of the request, also set on the response.
	ID string `json:"id"`
	// String to be uppercased.
	Input string `json:"input"`
	// Required information on which activity to send response to.
	ResponseActivity  string `json:"response_activity"`
	ResponseTaskQueue string `json:"response_task_queue"`
}

// Response is a response to a Request. This is the parameter of the response
// activity.
type Response struct {
	ID     string `json:"id"`
	Output string `json:"output"`
	Error  string `json:"error"`
}

type uppercaser struct {
	workflow.Context
	requestCh                   workflow.ReceiveChannel
	requestsBeforeContinueAsNew int
	responseActivityOptions     workflow.ActivityOptions
}

func newUppercaser(ctx workflow.Context) (*uppercaser, error) {
	u := &uppercaser{
		// For the main context, we're only going to allow 1 retry and only a 5
		// second schedule-to-close timeout. WARNING: The timeout and retry affect
		// how long this workflow stays open and may prevent it from performing its
		// continue-as-new until timeout occurs and/or retries are finished.
		Context: workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
			ScheduleToCloseTimeout: 5 * time.Second,
			RetryPolicy:            &temporal.RetryPolicy{MaximumAttempts: 2},
		}),

		// Use a signal to obtain requests. Since this cannot be closed, the
		// workflow counts on the ability to have some "idle" period between
		// handling requests/responses where it can continue-as-new.
		requestCh: workflow.GetSignalChannel(ctx, "request"),

		// We'll allow 500 requests before we continue-as-new the workflow. This is
		// required because the history will grow very large otherwise for an
		// interminable workflow fielding signal requests and executing activities.
		requestsBeforeContinueAsNew: 500,

		// We're gonna just support 10 second timeout and 3 retries on response
		// callback activities. WARNING: The timeout and retry affect how long this
		// workflow stays open and may prevent it from performing its
		// continue-as-new until timeout occurs and/or retries are finished.
		responseActivityOptions: workflow.ActivityOptions{
			// We use schedule-to-start/close because if the requester side is not
			// present, this may hang otherwise with just start-to-close.
			ScheduleToStartTimeout: 5 * time.Second,
			ScheduleToCloseTimeout: 10 * time.Second,
			RetryPolicy:            &temporal.RetryPolicy{MaximumAttempts: 4},
		},
	}
	return u, nil
}

func (u *uppercaser) run() error {
	// We create one selector for use for the entire run. Since we do a
	// continue-as-new after the workflow gets too large, continually adding
	// futures is ok.
	selector := workflow.NewSelector(u)

	// Listen for cancelled
	var cancelled bool
	selector.AddReceive(u.Done(), func(workflow.ReceiveChannel, bool) { cancelled = true })

	// Listen for new requests
	var requestCount int
	selector.AddReceive(u.requestCh, func(c workflow.ReceiveChannel, more bool) {
		requestCount++
		var req Request
		c.Receive(u, &req)
		u.addExecuteActivityFuture(&req, selector)
	})

	// Continually select until there are too many requests and no pending
	// selects.
	//
	// The reason we check selector.HasPending even when we've reached the request
	// limit is to make sure no events get lost. HasPending will continually
	// return true while an unresolved future or a buffered signal exists. If, for
	// example, we did not check this and there was an unhandled signal buffered
	// locally, continue-as-new would be returned without it being handled and the
	// new workflow wouldn't get the signal either. So it'd be lost.
	for requestCount < u.requestsBeforeContinueAsNew || selector.HasPending() {
		selector.Select(u)
		if cancelled {
			return temporal.NewCanceledError()
		}
	}

	// Continue as new since there were too many responses and the selector has
	// nothing pending. Note, if there is request signals come in faster than they
	// are handled or pending, there will not be a moment where the selector has
	// nothing pending which means this will run forever.
	return workflow.NewContinueAsNewError(u, UppercaseWorkflow)
}

func (u *uppercaser) addExecuteActivityFuture(req *Request, selector workflow.Selector) {
	// Add future for handling the result
	selector.AddFuture(workflow.ExecuteActivity(u, UppercaseActivity, req.Input), func(f workflow.Future) {
		resp := &Response{ID: req.ID}
		if err := f.Get(u, &resp.Output); err != nil {
			resp.Error = err.Error()
		}

		// Shallow copy activity options and set the task queue
		opts := u.responseActivityOptions
		opts.TaskQueue = req.ResponseTaskQueue
		actCtx := workflow.WithActivityOptions(u, opts)

		// We need to capture the error of the activity so we can log it. We add
		// the future to the selector instead of just doing a workflow.Go so that
		// we make sure the future is drained before continue-as-new occurs.
		// Otherwise, the future could be lost and the log may not occur.
		//
		// Note however that this ties the the workflow to the success/fail of
		// these response activities. Therefore, if these take longer to be
		// handled than the gap that may be needed between requests for
		// continue-as-new, the workflow will never exit.
		selector.AddFuture(workflow.ExecuteActivity(actCtx, req.ResponseActivity, resp), func(f workflow.Future) {
			// Just log if there is an error
			if err := f.Get(actCtx, nil); err != nil {
				workflow.GetLogger(actCtx).Warn("Failure sending response activity", "error", err)
			}
		})
	})
}
