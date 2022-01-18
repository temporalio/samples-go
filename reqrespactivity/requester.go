package reqrespactivity

import (
	"context"
	"fmt"
	"sync"

	"github.com/pborman/uuid"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

// Requester can request uppercasing of strings and should be closed after use.
type Requester struct {
	options   RequesterOptions
	taskQueue string
	// Channels need buffer of 1 because they are sent to in non-blocking fashion
	pendingRequests     map[string]chan<- *Response
	pendingRequestsLock sync.RWMutex
}

// RequesterOptions are options for NewRequester.
type RequesterOptions struct {
	// Client to the Temporal server. Required. Not closed on Requester.Close.
	Client client.Client
	// ID of the workflow listening for signals to uppercase. Required.
	TargetWorkflowID string

	// Visible for testing.
	ExistingWorker interface {
		RegisterActivity(interface{})
		Start() error
		Stop()
	}
}

// NewRequester creates a new Requester for the given options.
func NewRequester(options RequesterOptions) (*Requester, error) {
	if options.Client == nil {
		return nil, fmt.Errorf("client required")
	} else if options.TargetWorkflowID == "" {
		return nil, fmt.Errorf("target workflow required")
	}

	// Create requester
	r := &Requester{
		options:         options,
		taskQueue:       "requester-" + uuid.New(),
		pendingRequests: map[string]chan<- *Response{},
	}
	if r.options.ExistingWorker == nil {
		r.options.ExistingWorker = worker.New(options.Client, r.taskQueue, worker.Options{})
	}

	// Start worker
	r.options.ExistingWorker.RegisterActivity(r.responseActivity)
	if err := r.options.ExistingWorker.Start(); err != nil {
		return nil, fmt.Errorf("failed starting worker: %w", err)
	}
	return r, nil
}

// RequestUppercase sends a request and returns a response.
func (r *Requester) RequestUppercase(ctx context.Context, str string) (string, error) {
	// Create request and add channel to pending
	req := &Request{
		ID:                uuid.New(),
		Input:             str,
		ResponseActivity:  "responseActivity",
		ResponseTaskQueue: r.taskQueue,
	}
	respCh := make(chan *Response, 1)
	r.pendingRequestsLock.Lock()
	r.pendingRequests[req.ID] = respCh
	r.pendingRequestsLock.Unlock()

	// Remove pending request when done
	defer func() {
		r.pendingRequestsLock.Lock()
		defer r.pendingRequestsLock.Unlock()
		delete(r.pendingRequests, req.ID)
	}()

	// Send request and wait for response
	if err := r.options.Client.SignalWorkflow(ctx, r.options.TargetWorkflowID, "", "request", req); err != nil {
		return "", fmt.Errorf("failed signaling workflow: %w", err)
	}
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case resp := <-respCh:
		if resp.Error != "" {
			return "", fmt.Errorf("Request failed: %v", resp.Error)
		}
		return resp.Output, nil
	}
}

// Close stops the internal worker. Since this stops the response worker, it
// does a graceful stop for a period. Callers are expected to not call requests
// after this and to cancel outstanding requests.
func (r *Requester) Close() {
	// We intentionally don't cancel all pending requests because they may
	// complete during graceful worker stop. We expect the caller to no longer
	// call for new requests and cancel their request contexts.
	r.options.ExistingWorker.Stop()
}

func (r *Requester) responseActivity(ctx context.Context, resp *Response) error {
	// Get the channel to respond to
	r.pendingRequestsLock.RLock()
	respCh := r.pendingRequests[resp.ID]
	r.pendingRequestsLock.RUnlock()
	// We choose not to log or error if a response is not pending because it is
	// normal behavior for a requester to have closed the context and stop waiting
	if respCh == nil {
		return nil
	}

	// Send non-blocking since the channel should have enough room. Technically
	// during a situation where this worker was too busy for this activity to
	// return, the responseActivity could be called again for the same response
	// during retry from the other side. This will just result in a no-op since
	// the channel does not have room.
	select {
	case respCh <- resp:
	default:
	}
	return nil
}
