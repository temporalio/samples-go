package reqrespactivity_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/temporalio/samples-go/reqrespactivity"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
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

func TestUppercaseWorkflowWaitsForPendingResponsesBeforeContinueAsNew(t *testing.T) {
	// Create env
	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestWorkflowEnvironment()
	env.RegisterWorkflow(reqrespactivity.UppercaseWorkflow)
	env.RegisterActivity(reqrespactivity.UppercaseActivity)

	// Add the external activity. Activities can run concurrently, so guard the
	// responses with a mutex.
	var externalResponsesLock sync.Mutex
	var externalResponses []*reqrespactivity.Response
	env.RegisterActivityWithOptions(
		func(resp *reqrespactivity.Response) error {
			externalResponsesLock.Lock()
			defer externalResponsesLock.Unlock()
			externalResponses = append(externalResponses, resp)
			return nil
		},
		activity.RegisterOptions{Name: "external-activity"},
	)

	// Make the uppercase activity take a while so requests are still in flight
	// when the request limit is reached
	env.OnActivity(reqrespactivity.UppercaseActivity, mock.Anything, mock.Anything).
		After(time.Second).
		Return(func(_ context.Context, input string) (string, error) {
			return reqrespactivity.UppercaseActivity(context.Background(), input)
		})

	// Send enough requests to reach the continue-as-new limit
	const requestCount = 500
	env.RegisterDelayedCallback(func() {
		for i := 0; i < requestCount; i++ {
			env.SignalWorkflow("request", &reqrespactivity.Request{
				ID:                fmt.Sprintf("request%d", i),
				Input:             "foo",
				ResponseActivity:  "external-activity",
				ResponseTaskQueue: "external-task-queue",
			})
		}
	}, 2)

	// Run workflow
	env.ExecuteWorkflow(reqrespactivity.UppercaseWorkflow)

	// Confirm the workflow continued as new only after every response was sent
	require.True(t, env.IsWorkflowCompleted())
	var continueAsNewErr *workflow.ContinueAsNewError
	require.ErrorAs(t, env.GetWorkflowError(), &continueAsNewErr)
	require.Len(t, externalResponses, requestCount)
}
