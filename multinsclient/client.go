package multinsclient

import (
	"bytes"
	"context"
	"fmt"
	"hash/fnv"

	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
)

// Namespace represents a client and a namespace for connecting.
type Namespace struct {
	// Options for the client. The Namespace field must be populated. The options
	// are never used to connect to a client if Disabled is true.
	ClientOptions client.Options
	// If true, this namespace will never have a client created for it.
	Disabled bool
}

// Client contains clients for different namespaces.
type Client struct {
	allNamespaces           []string
	enabledNamespaces       []string
	enabledNamespaceClients map[string]client.Client
	hasher                  func(workflowID string, max int) int
}

// Options for New.
type Options struct {
	// Namespaces this client will work with. The order and count and names should
	// remain the same always so the hashing is somewhat consistent. At least one
	// enabled namespace is required.
	//
	// See Client.NamespaceFor for details on namespace choice.
	Namespaces []Namespace

	// Hasher that chooses a number from 0 (inclusive) to max (exclusive) for the
	// given workflow ID. If unset, DefaultHasher is used.
	Hasher func(workflowID string, max int) int

	// If true, enabled namespaces will create clients that won't connect until
	// first used. Otherwise, enabled namespaces will have their clients eagerly
	// attempt to connect to the server when the client is created.
	ConnectLazy bool
}

// DefaultHasher uses FNV-1a hashing.
func DefaultHasher(workflowID string, max int) int {
	h := fnv.New32a()
	h.Write([]byte(workflowID))
	return int(h.Sum32()) % max
}

// New creates a new client from the given options.
func New(options Options) (*Client, error) {
	c := &Client{
		hasher:                  options.Hasher,
		enabledNamespaceClients: map[string]client.Client{},
	}

	// Use default hasher if none given
	if c.hasher == nil {
		c.hasher = DefaultHasher
	}

	// Close any created clients of this didn't complete successfully
	success := false
	defer func() {
		if !success {
			for _, client := range c.enabledNamespaceClients {
				client.Close()
			}
		}
	}()

	// Create all non-disabled clients
	seenNamespaces := make(map[string]bool, len(options.Namespaces))
	for _, namespace := range options.Namespaces {
		// We don't support default namespace name fallback, we require it to be set
		if namespace.ClientOptions.Namespace == "" {
			return nil, fmt.Errorf("all namespaces must have a namespace in the client options")
		}
		if _, exists := seenNamespaces[namespace.ClientOptions.Namespace]; exists {
			return nil, fmt.Errorf("duplicate namespace %q", namespace.ClientOptions.Namespace)
		}
		seenNamespaces[namespace.ClientOptions.Namespace] = true
		c.allNamespaces = append(c.allNamespaces, namespace.ClientOptions.Namespace)
		if !namespace.Disabled {
			c.enabledNamespaces = append(c.enabledNamespaces, namespace.ClientOptions.Namespace)
			var err error
			// TODO(cretz): Could alter options to have client options unrelated to
			// namespaces and then use the new client.NewClientFromExisting introduced
			// in https://github.com/temporalio/sdk-go/pull/881 to share connections
			// across namespaces, but having multiple clients is mostly harmless.
			if options.ConnectLazy {
				c.enabledNamespaceClients[namespace.ClientOptions.Namespace], err = client.NewLazyClient(namespace.ClientOptions)
			} else {
				c.enabledNamespaceClients[namespace.ClientOptions.Namespace], err = client.Dial(namespace.ClientOptions)
			}
			if err != nil {
				return nil, fmt.Errorf("failed creating client for namespace %q: %w", namespace.ClientOptions.Namespace, err)
			}
		}
	}

	// Must be at least one enabled
	if len(c.enabledNamespaces) == 0 {
		return nil, fmt.Errorf("no enabled namespaces")
	}

	success = true
	return c, nil
}

// NamespaceFor selects a namespace and client for the given workflow ID. The
// resulting namespace is always the same for the same workflow ID.
//
// If onlyEnabled is true and the namespace that would be selected is not
// enabled, a hash is done only across the enabled namespace. Therefore, while
// the same namespace is always selected even after Reset for the same workflow
// ID when onlyEnabled is false, when onlyEnabled is true, the selected
// namespace may be different for each Reset.
//
// The resulting client is only non-nil if the namespace is enabled.
func (c *Client) NamespaceFor(workflowID string, onlyEnabled bool) (namespace string, client client.Client) {
	if workflowID == "" {
		panic("missing workflow ID")
	}
	// First try across all namespaces
	namespace = c.allNamespaces[c.hasher(workflowID, len(c.allNamespaces))]
	// If they only want enabled and this one isn't enabled, do another hash only
	// across the enabled ones
	if onlyEnabled && c.enabledNamespaceClients[namespace] == nil {
		namespace = c.enabledNamespaces[c.hasher(workflowID, len(c.enabledNamespaces))]
	}
	client = c.enabledNamespaceClients[namespace]
	return
}

// ClientFor is a shortcut for NamespaceFor with onlyEnabled as true.
func (c *Client) ClientFor(workflowID string) client.Client {
	_, client := c.NamespaceFor(workflowID, true)
	return client
}

// Reset creates a new client from the given options. The namespaces must be the
// same count with the same names in the same order. If closeCurrent is true,
// this will close the current client before returning a new client on success.
//
// Reset is used for creating a new client if which namespaces are enabled
// changes. This does close and reconnect all client connections.
func (c *Client) Reset(options Options, closeCurrent bool) (*Client, error) {
	// The options must have the same number of namespaces in the same order
	if len(c.allNamespaces) != len(options.Namespaces) {
		return nil, fmt.Errorf("cannot reset with different number of namespaces")
	}

	// Confirm no namespace names changed
	for i, namespace := range options.Namespaces {
		if c.allNamespaces[i] != namespace.ClientOptions.Namespace {
			return nil, fmt.Errorf(
				"existing namespace at index %v of %q different than new of %q (must be same names in same order)",
				i, c.allNamespaces[i], namespace.ClientOptions.Namespace)
		}
	}

	// Create client then maybe close this one
	newClient, err := New(options)
	if err == nil && closeCurrent {
		c.Close()
	}
	return newClient, err
}

// Close closes all clients of enabled namespaces. No uses of clients should
// occur after this is called.
func (c *Client) Close() {
	for _, client := range c.enabledNamespaceClients {
		client.Close()
	}
}

// ExecuteWorkflow is a multi-namespace version of client.ExecuteWorkflow. The
// workflow ID must be present.
func (c *Client) ExecuteWorkflow(
	ctx context.Context,
	options client.StartWorkflowOptions,
	workflow interface{},
	args ...interface{},
) (client.WorkflowRun, error) {
	if options.ID == "" {
		return nil, fmt.Errorf("workflow ID must be present")
	}
	return c.ClientFor(options.ID).ExecuteWorkflow(ctx, options, workflow, args...)
}

// GetWorkflow is a multi-namespace version of client.GetWorkflow.
func (c *Client) GetWorkflow(ctx context.Context, workflowID string, runID string) client.WorkflowRun {
	return c.ClientFor(workflowID).GetWorkflow(ctx, workflowID, runID)
}

// SignalWorkflow is a multi-namespace version of client.SignalWorkflow.
func (c *Client) SignalWorkflow(
	ctx context.Context,
	workflowID string,
	runID string,
	signalName string,
	arg interface{},
) error {
	return c.ClientFor(workflowID).SignalWorkflow(ctx, workflowID, runID, signalName, arg)
}

// SignalWithStartWorkflow is a multi-namespace version of
// client.SignalWithStartWorkflow.
func (c *Client) SignalWithStartWorkflow(
	ctx context.Context,
	workflowID string,
	signalName string,
	signalArg interface{},
	options client.StartWorkflowOptions,
	workflow interface{},
	workflowArgs ...interface{},
) (client.WorkflowRun, error) {
	return c.ClientFor(workflowID).SignalWithStartWorkflow(
		ctx, workflowID, signalName, signalArg, options, workflow, workflowArgs...)
}

// CancelWorkflow is a multi-namespace version of client.CancelWorkflow.
func (c *Client) CancelWorkflow(ctx context.Context, workflowID string, runID string) error {
	return c.ClientFor(workflowID).CancelWorkflow(ctx, workflowID, runID)
}

// TerminateWorkflow is a multi-namespace version of client.TerminateWorkflow.
func (c *Client) TerminateWorkflow(
	ctx context.Context,
	workflowID string,
	runID string,
	reason string,
	details ...interface{},
) error {
	return c.ClientFor(workflowID).TerminateWorkflow(ctx, workflowID, runID, reason, details...)
}

// ListWorkflow is a multi-namespace version of client.ListWorkflow. The page
// size on the request must be set and namespace cannot be set.
func (c *Client) ListWorkflow(
	ctx context.Context,
	request *workflowservice.ListWorkflowExecutionsRequest,
) (*workflowservice.ListWorkflowExecutionsResponse, error) {
	if request.Namespace != "" {
		return nil, fmt.Errorf("namespace not allowed on request")
	} else if request.PageSize == 0 {
		// TODO(cretz): Default page size?
		return nil, fmt.Errorf("page size required")
	}

	// If there is a next-page-token, extract the starting namespace
	var startingNamespace string
	callReq := &workflowservice.ListWorkflowExecutionsRequest{Query: request.Query}
	if len(request.NextPageToken) > 0 {
		callReq.NextPageToken = bytes.TrimPrefix(request.NextPageToken, []byte("__nspre__"))
		if len(callReq.NextPageToken) == len(request.NextPageToken) {
			return nil, fmt.Errorf("invalid next page token")
		}
		postIndex := bytes.Index(callReq.NextPageToken, []byte("__nspost__"))
		if postIndex == -1 {
			return nil, fmt.Errorf("invalid next page token")
		}
		startingNamespace = string(callReq.NextPageToken[:postIndex])
		callReq.NextPageToken = bytes.TrimPrefix(callReq.NextPageToken[postIndex:], []byte("__nspost__"))
	}

	// Continually retrieve pages until we fill a page or are done
	var resp workflowservice.ListWorkflowExecutionsResponse
	for i, namespace := range c.enabledNamespaces {
		// If there's a starting namespace and we're not it, keep going
		if startingNamespace != "" {
			if startingNamespace != namespace {
				continue
			}
			startingNamespace = ""
		}

		// Set the page size as how many left we need
		callReq.PageSize = request.PageSize - int32(len(resp.Executions))
		callReq.Namespace = namespace
		callResp, err := c.enabledNamespaceClients[namespace].ListWorkflow(ctx, callReq)
		if err != nil {
			return nil, err
		}

		// Add executions and if we've reached the page limit we're done
		resp.Executions = append(resp.Executions, callResp.Executions...)
		if len(resp.Executions) == int(request.PageSize) {
			// If there is a next page token, set it. Otherwise, if we're not at the
			// last namespace, set the next one as starting with no token.
			if len(callResp.NextPageToken) > 0 {
				resp.NextPageToken = append([]byte("__nspre__"+namespace+"__nspost__"), callResp.NextPageToken...)
			} else if i < len(c.enabledNamespaces)-1 {
				resp.NextPageToken = []byte("__nspre__" + c.enabledNamespaces[i+1] + "__nspost__")
			}
			break
		}
		// Clear out next token before doing next namespace
		callReq.NextPageToken = nil
	}
	return &resp, nil
}
