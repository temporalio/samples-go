package multinsclient

import (
	"fmt"
	"hash/fnv"

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
