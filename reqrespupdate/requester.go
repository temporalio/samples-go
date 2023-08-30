package reqrespupdate

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/client"
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
}

// NewRequester creates a new Requester for the given options.
func NewRequester(options RequesterOptions) (*Requester, error) {
	if options.Client == nil {
		return nil, fmt.Errorf("client required")
	} else if options.TargetWorkflowID == "" {
		return nil, fmt.Errorf("target workflow required")
	}
	return &Requester{options}, nil
}

// RequestUppercase sends a request and returns a response.
func (r *Requester) RequestUppercase(ctx context.Context, str string) (string, error) {
	// Send request and poll on an interval for response
	handle, err := r.options.Client.UpdateWorkflow(ctx, r.options.TargetWorkflowID, "", UpdateHandler, Request{Input: str})
	if err != nil {
		return "", fmt.Errorf("failed updating workflow: %w", err)
	}
	var response Response
	err = handle.Get(ctx, &response)
	if err != nil {
		return "", fmt.Errorf("failed getting update response: %w", err)
	}
	return response.Output, nil
}
