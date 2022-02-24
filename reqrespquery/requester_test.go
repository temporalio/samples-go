package reqrespquery_test

import (
	"context"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/temporalio/samples-go/reqrespquery"
	"go.temporal.io/sdk/mocks"
)

func TestRequester(t *testing.T) {
	c := &mocks.Client{}
	// Handle query requests
	var queryResponse *reqrespquery.Response
	var queryResponseLock sync.RWMutex
	queryVal := &mocks.Value{}
	queryVal.On("Get", mock.AnythingOfType("**reqrespquery.Response")).
		Maybe().
		Return(nil).
		Run(func(args mock.Arguments) {
			queryResponseLock.RLock()
			defer queryResponseLock.RUnlock()
			*args.Get(0).(**reqrespquery.Response) = queryResponse
		})
	c.On("QueryWorkflow", mock.Anything, "some-workflow", "", "response", mock.AnythingOfType("string")).
		Maybe().
		Return(queryVal, nil)
	// Expect to be signalled
	c.On("SignalWorkflow", mock.Anything, "some-workflow", "", "request", mock.AnythingOfType("*reqrespquery.Request")).
		Once().
		Return(nil).
		Run(func(args mock.Arguments) {
			// Once signalled, we can now respond to the query
			queryResponseLock.Lock()
			defer queryResponseLock.Unlock()
			req := args[len(args)-1].(*reqrespquery.Request)
			queryResponse = &reqrespquery.Response{ID: req.ID, Output: strings.ToUpper(req.Input)}
		})

	// Create requester
	req, err := reqrespquery.NewRequester(reqrespquery.RequesterOptions{
		Client:           c,
		TargetWorkflowID: "some-workflow",
	})
	require.NoError(t, err)

	// Request
	res, err := req.RequestUppercase(context.Background(), "SoMe VaLuE")
	require.NoError(t, err)
	require.Equal(t, "SOME VALUE", res)
}
