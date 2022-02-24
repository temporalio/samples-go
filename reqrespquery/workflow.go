package reqrespquery

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// UppercaseWorkflow is a workflow that accepts requests to uppercase strings
// via signals and provides responses via query.
//
// The "request" signal accepts a Request.
//
// The "response" query accepts a request ID.
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
}

// Response is a response to a Request. This can be returned in a map from a
// query or as the parameter of a callback to a response activity.
type Response struct {
	ID     string `json:"id"`
	Output string `json:"output"`
	Error  string `json:"error"`
}

type uppercaser struct {
	workflow.Context
	requestCh                   workflow.ReceiveChannel
	requestsBeforeContinueAsNew int
	// This maintains responses for the lifetime of the workflow, which should not
	// be too large before continue-as-new.
	responses map[string]*Response
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
		responses:                   make(map[string]*Response, 500),
	}
	// Set query handler
	if err := workflow.SetQueryHandler(ctx, "response", u.queryResponse); err != nil {
		return nil, fmt.Errorf("failed setting query handler: %w", err)
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

func (u *uppercaser) queryResponse(id string) (*Response, error) {
	// We intentionally do not remove the response from the map for the
	// following reasons:
	// * It is not acceptable for a query to have side effects
	// * The query may be called multiple times by the requester even when found
	// * A workflow writer should treat their callers as external entities and
	//   should not assume how the query will be used
	return u.responses[id], nil
}

func (u *uppercaser) addExecuteActivityFuture(req *Request, selector workflow.Selector) {
	// Add future for handling the result
	selector.AddFuture(workflow.ExecuteActivity(u, UppercaseActivity, req.Input), func(f workflow.Future) {
		resp := &Response{ID: req.ID}
		if err := f.Get(u, &resp.Output); err != nil {
			resp.Error = err.Error()
		}
		u.responses[resp.ID] = resp
	})
}
