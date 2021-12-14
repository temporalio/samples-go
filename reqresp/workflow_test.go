package reqresp

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/testsuite"
)

func TestUppercaseWorkflow(t *testing.T) {
	// Create env
	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestWorkflowEnvironment()
	env.RegisterWorkflow(UppercaseWorkflow)
	env.RegisterActivity(UppercaseActivity)

	// Add a delayed callback to send a signal without activity response
	env.RegisterDelayedCallback(func() { env.SignalWorkflow("request", &Request{ID: "request1", Input: "foo"}) }, 1)

	// Add another delayed callback to send a signal with activity response
	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow("request", &Request{
			ID:                "request2",
			Input:             "bar",
			ResponseActivity:  "external-activity",
			ResponseTaskQueue: "external-task-queue",
		})
	}, 2)

	// Add a delayed callback after that to query
	var queryResponses map[string]*Response
	var queryErr error
	env.RegisterDelayedCallback(func() {
		val, err := env.QueryWorkflow("response", []string{"request1"})
		if err != nil {
			queryErr = err
		} else {
			queryErr = val.Get(&queryResponses)
		}
	}, 100*time.Millisecond)

	// Add the external activity
	var externalResponses []*Response
	env.RegisterActivityWithOptions(
		func(resp *Response) error {
			externalResponses = append(externalResponses, resp)
			return nil
		},
		activity.RegisterOptions{Name: "external-activity"},
	)

	// Run workflow
	env.ExecuteWorkflow(UppercaseWorkflow)

	// Confirm query response
	require.NoError(t, queryErr)
	require.Equal(t, map[string]*Response{"request1": {ID: "request1", Output: "FOO"}}, queryResponses)

	// Confirm activity sent
	require.Equal(t, []*Response{{ID: "request2", Output: "BAR"}}, externalResponses)
}
