package reqrespactivity_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/temporalio/samples-go/reqrespactivity"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/testsuite"
)

func TestUppercaseWorkflow(t *testing.T) {
	// Create env
	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestWorkflowEnvironment()
	env.RegisterWorkflow(reqrespactivity.UppercaseWorkflow)
	env.RegisterActivity(reqrespactivity.UppercaseActivity)

	// Add another delayed callback to send a signal
	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow("request", &reqrespactivity.Request{
			ID:                "request1",
			Input:             "foo",
			ResponseActivity:  "external-activity",
			ResponseTaskQueue: "external-task-queue",
		})
	}, 2)

	// Add the external activity
	var externalResponses []*reqrespactivity.Response
	env.RegisterActivityWithOptions(
		func(resp *reqrespactivity.Response) error {
			externalResponses = append(externalResponses, resp)
			return nil
		},
		activity.RegisterOptions{Name: "external-activity"},
	)

	// Run workflow
	env.ExecuteWorkflow(reqrespactivity.UppercaseWorkflow)

	// Confirm activity sent
	require.Equal(t, []*reqrespactivity.Response{{ID: "request1", Output: "FOO"}}, externalResponses)
}
