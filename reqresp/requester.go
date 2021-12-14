package reqresp

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/pborman/uuid"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

// Requester can request uppercasing of strings and should be closed after use.
type Requester interface {
	// RequestUppercase requests uppercasing the given string.
	RequestUppercase(ctx context.Context, str string) (string, error)

	// Close closes this requester.
	Close()
}

// RequesterOptions are options for NewRequester.
type RequesterOptions struct {
	// Client to the Temporal server. Required. Not closed on Requester.Close.
	Client client.Client
	// ID of the workflow listening for signals to uppercase. Required.
	TargetWorkflowID string
	// If true, this uses activity-based response handling which is more robust
	// than the query-based polling approach done when this is false.
	UseActivityResponse bool
}

// NewRequester creates a new Requester for the given options.
func NewRequester(opts RequesterOptions) (Requester, error) {
	if opts.Client == nil {
		return nil, fmt.Errorf("client required")
	} else if opts.TargetWorkflowID == "" {
		return nil, fmt.Errorf("target workflow required")
	}
	if opts.UseActivityResponse {
		return newRequesterActivityResponses(opts)
	}
	return newRequesterQueryResponses(opts), nil
}

type requesterBase struct {
	RequesterOptions
	// Channels need buffer because they are sent to in non-blocking fashion
	pendingRequests     map[string]chan<- *Response
	pendingRequestsLock sync.RWMutex
}

func (r *requesterBase) requestUppercase(ctx context.Context, req *Request) (string, error) {
	// Create request and add channel to pending
	r.pendingRequestsLock.Lock()
	respCh := make(chan *Response, 1)
	r.pendingRequests[req.ID] = respCh
	r.pendingRequestsLock.Unlock()

	// Remove pending request when done
	defer func() {
		r.pendingRequestsLock.Lock()
		defer r.pendingRequestsLock.Unlock()
		delete(r.pendingRequests, req.ID)
	}()

	// Send and signal wait for response
	if err := r.Client.SignalWorkflow(ctx, r.TargetWorkflowID, "", "request", req); err != nil {
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

type requesterActivityResponses struct {
	requesterBase
	taskQueue string
	worker    activityWorker
}

// Visible for testing
type activityWorker interface {
	RegisterActivity(interface{})
	Start() error
	Stop()
}

var newWorker = func(c client.Client, taskQueue string) activityWorker {
	return worker.New(c, taskQueue, worker.Options{})
}

func newRequesterActivityResponses(opts RequesterOptions) (*requesterActivityResponses, error) {
	var r requesterActivityResponses
	r.RequesterOptions = opts
	r.pendingRequests = map[string]chan<- *Response{}
	r.taskQueue = "requester-" + uuid.New()
	r.worker = newWorker(opts.Client, r.taskQueue)
	// Register the activity and start
	r.worker.RegisterActivity(r.responseActivity)
	if err := r.worker.Start(); err != nil {
		return nil, fmt.Errorf("failed starting worker: %w", err)
	}
	return &r, nil
}

func (r *requesterActivityResponses) RequestUppercase(ctx context.Context, str string) (string, error) {
	// Send request to use the response activity
	return r.requestUppercase(ctx,
		&Request{ID: uuid.New(), Input: str, ResponseActivity: "responseActivity", ResponseTaskQueue: r.taskQueue})
}

func (r *requesterActivityResponses) Close() { r.worker.Stop() }

func (r *requesterActivityResponses) responseActivity(ctx context.Context, resp *Response) error {
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

// Visible for testing
var tickerFreq = 3 * time.Second

type requesterQueryResponses struct {
	requesterBase
	ticker     *time.Ticker
	pollCtx    context.Context
	pollCancel context.CancelFunc
}

func newRequesterQueryResponses(opts RequesterOptions) *requesterQueryResponses {
	var r requesterQueryResponses
	r.RequesterOptions = opts
	r.pendingRequests = map[string]chan<- *Response{}
	// Query every so often
	r.ticker = time.NewTicker(tickerFreq)
	r.pollCtx, r.pollCancel = context.WithCancel(context.Background())
	// Start poller
	go r.poll()
	return &r
}

func (r *requesterQueryResponses) RequestUppercase(ctx context.Context, str string) (string, error) {
	// Send request
	return r.requestUppercase(ctx, &Request{ID: uuid.New(), Input: str})
}

func (r *requesterQueryResponses) Close() {
	r.ticker.Stop()
	r.pollCancel()
}

func (r *requesterQueryResponses) poll() {
	for {
		select {
		case <-r.pollCtx.Done():
			return
		case <-r.ticker.C:
			// Collect all pending IDs
			r.pendingRequestsLock.RLock()
			pendingIDs := make([]string, 0, len(r.pendingRequests))
			pendingResponses := map[string]chan<- *Response{}
			for id, resp := range r.pendingRequests {
				pendingIDs = append(pendingIDs, id)
				pendingResponses[id] = resp
			}
			r.pendingRequestsLock.RUnlock()

			// Query for their responses and just log errors but don't exit poll
			val, err := r.Client.QueryWorkflow(r.pollCtx, r.TargetWorkflowID, "", "response", pendingIDs)
			if err != nil {
				// We don't want to report a stop
				if r.pollCtx.Err() == nil {
					log.Printf("Query workflow failed: %v", err)
				}
				continue
			}
			// Convert to response map and send in non-blocking fashion to each
			var responses map[string]*Response
			if err := val.Get(&responses); err != nil {
				log.Printf("Failed converting query response: %v", err)
				continue
			}
			for id, resp := range responses {
				select {
				case pendingResponses[id] <- resp:
				default:
				}
			}
		}
	}
}
