package reqrespquery_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/temporalio/samples-go/reqrespquery"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"
)

func TestUppercaseWorkflow(t *testing.T) {
	// Create env
	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestWorkflowEnvironment()
	env.RegisterWorkflow(reqrespquery.UppercaseWorkflow)
	env.RegisterActivity(reqrespquery.UppercaseActivity)

	// Add a delayed callback to send a signal
	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow("request", &reqrespquery.Request{ID: "request1", Input: "foo"})
	}, 1)

	// Add a delayed callback after that to query
	var queryResponse *reqrespquery.Response
	var queryErr error
	env.RegisterDelayedCallback(func() {
		val, err := env.QueryWorkflow("response", "request1")
		if err != nil {
			queryErr = err
		} else {
			queryErr = val.Get(&queryResponse)
		}
	}, 100*time.Millisecond)

	// Run workflow
	env.ExecuteWorkflow(reqrespquery.UppercaseWorkflow)

	// Confirm query response
	require.NoError(t, queryErr)
	require.Equal(t, &reqrespquery.Response{ID: "request1", Output: "FOO"}, queryResponse)
}

func TestUppercaseWorkflowWaitsForPendingResponsesBeforeContinueAsNew(t *testing.T) {
	// Create env
	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestWorkflowEnvironment()
	env.RegisterWorkflow(reqrespquery.UppercaseWorkflow)
	env.RegisterActivity(reqrespquery.UppercaseActivity)

	// Make the uppercase activity take a while so requests are still in flight
	// when the request limit is reached
	env.OnActivity(reqrespquery.UppercaseActivity, mock.Anything, mock.Anything).
		After(time.Second).
		Return(func(_ context.Context, input string) (string, error) {
			return reqrespquery.UppercaseActivity(context.Background(), input)
		})

	// Send enough requests to reach the continue-as-new limit
	const requestCount = 500
	env.RegisterDelayedCallback(func() {
		for i := 0; i < requestCount; i++ {
			env.SignalWorkflow("request", &reqrespquery.Request{ID: fmt.Sprintf("request%d", i), Input: "foo"})
		}
	}, 1)

	// Run workflow
	env.ExecuteWorkflow(reqrespquery.UppercaseWorkflow)

	// Confirm the workflow continued as new only after the last response was set
	require.True(t, env.IsWorkflowCompleted())
	var continueAsNewErr *workflow.ContinueAsNewError
	require.ErrorAs(t, env.GetWorkflowError(), &continueAsNewErr)
	val, err := env.QueryWorkflow("response", fmt.Sprintf("request%d", requestCount-1))
	require.NoError(t, err)
	var queryResponse *reqrespquery.Response
	require.NoError(t, val.Get(&queryResponse))
	require.Equal(t, &reqrespquery.Response{ID: fmt.Sprintf("request%d", requestCount-1), Output: "FOO"}, queryResponse)
}
