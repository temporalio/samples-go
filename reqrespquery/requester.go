package reqrespquery

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/pborman/uuid"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
)

// Requester can request uppercasing of strings.
type Requester struct {
	options RequesterOptions
}

// RequesterOptions are options for NewRequester.
type RequesterOptions struct {
	// Client to the Temporal server. Required.
	Client client.Client
	// ID of the workflow listening for signals to uppercase. Required.
	TargetWorkflowID string
	// Frequency of query for response. Default 300ms.
	ResponseQueryInterval time.Duration
	// How long to wait for response. Default 4 seconds.
	ResponseTimeout time.Duration
}

// NewRequester creates a new Requester for the given options.
func NewRequester(options RequesterOptions) (*Requester, error) {
	if options.Client == nil {
		return nil, fmt.Errorf("client required")
	} else if options.TargetWorkflowID == "" {
		return nil, fmt.Errorf("target workflow required")
	}
	if options.ResponseQueryInterval == 0 {
		options.ResponseQueryInterval = 300 * time.Millisecond
	}
	if options.ResponseTimeout == 0 {
		options.ResponseTimeout = 4 * time.Second
	}
	return &Requester{options}, nil
}

// RequestUppercase sends a request and returns a response.
func (r *Requester) RequestUppercase(ctx context.Context, str string) (string, error) {
	// Send request and poll on an interval for response
	req := &Request{ID: uuid.New(), Input: str}
	if err := r.options.Client.SignalWorkflow(ctx, r.options.TargetWorkflowID, "", "request", req); err != nil {
		return "", fmt.Errorf("failed signaling workflow: %w", err)
	}

	t := time.NewTicker(r.options.ResponseQueryInterval)
	defer t.Stop()
	pollCtx, cancel := context.WithTimeout(ctx, r.options.ResponseTimeout)
	defer cancel()

	for {
		// Query for response
		val, err := r.options.Client.QueryWorkflow(pollCtx, r.options.TargetWorkflowID, "", "response", req.ID)

		// Sometimes an error can happen during continue-as-new of a workflow, so we
		// do not fail on errors here, we just log them
		if err != nil {
			log.Printf("Query workflow failed: %v", err)
		} else if val != nil {
			var resp *Response
			err = val.Get(&resp)
			// We continue on ErrNoData
			if err != nil && err != temporal.ErrNoData {
				return "", fmt.Errorf("failed unmarshalling response: %w", err)
			} else if resp != nil && resp.Error != "" {
				return "", fmt.Errorf("request failed: %v", resp.Error)
			} else if resp != nil {
				return resp.Output, nil
			}
		}

		// Wait for interval or timeout
		select {
		case <-pollCtx.Done():
			// If we timed out and not the original context, show last error
			if ctx.Err() == nil && err != nil {
				return "", fmt.Errorf("timeout, last error: %w", err)
			}
			return "", ctx.Err()
		case <-t.C:
		}
	}
}
