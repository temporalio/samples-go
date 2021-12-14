package reqresp

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/mocks"
	"go.temporal.io/sdk/testsuite"
)

func TestActivityRequester(t *testing.T) {
	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestActivityEnvironment()

	// Change worker constructor for the test
	newWorker = func(client.Client, string) activityWorker { return &fakeWorker{env} }

	// Expect to be signalled
	c := &mocks.Client{}
	c.On("SignalWorkflow", mock.Anything, "some-workflow", "", "request", mock.AnythingOfType("*reqresp.Request")).
		Once().
		Return(nil).
		Run(func(args mock.Arguments) {
			// Once signalled, we can now execute the activity
			req := args[len(args)-1].(*Request)
			_, err := env.ExecuteActivity(req.ResponseActivity, &Response{
				ID:     req.ID,
				Output: strings.ToUpper(req.Input),
			})
			require.NoError(t, err)
		})

	// Create requester
	req, err := NewRequester(RequesterOptions{
		Client:              c,
		TargetWorkflowID:    "some-workflow",
		UseActivityResponse: true,
	})
	require.NoError(t, err)
	defer req.Close()

	// Request
	res, err := req.RequestUppercase(context.Background(), "SoMe VaLuE")
	require.NoError(t, err)
	require.Equal(t, "SOME VALUE", res)
}

type fakeWorker struct {
	env *testsuite.TestActivityEnvironment
}

func (f *fakeWorker) RegisterActivity(a interface{}) { f.env.RegisterActivity(a) }
func (*fakeWorker) Start() error                     { return nil }
func (*fakeWorker) Stop()                            {}

func TestQueryRequester(t *testing.T) {
	// Tick much more frequently
	tickerFreq = 100 * time.Millisecond

	c := &mocks.Client{}
	// Handle query requests
	queryResponses := map[string]*Response{}
	var queryResponsesLock sync.RWMutex
	queryVal := &mocks.Value{}
	queryVal.On("Get", mock.AnythingOfType("*map[string]*reqresp.Response")).
		Maybe().
		Return(nil).
		Run(func(args mock.Arguments) {
			queryResponsesLock.RLock()
			defer queryResponsesLock.RUnlock()
			arg := args.Get(0).(*map[string]*Response)
			*arg = map[string]*Response{}
			for k, v := range queryResponses {
				(*arg)[k] = v
			}
		})
	c.On("QueryWorkflow", mock.Anything, "some-workflow", "", "response", mock.AnythingOfType("[]string")).
		Maybe().
		Return(queryVal, nil)
	// Expect to be signalled
	c.On("SignalWorkflow", mock.Anything, "some-workflow", "", "request", mock.AnythingOfType("*reqresp.Request")).
		Once().
		Return(nil).
		Run(func(args mock.Arguments) {
			// Once signalled, we can now respond to the query
			queryResponsesLock.Lock()
			defer queryResponsesLock.Unlock()
			req := args[len(args)-1].(*Request)
			queryResponses[req.ID] = &Response{
				ID:     req.ID,
				Output: strings.ToUpper(req.Input),
			}
		})

	// Create requester
	req, err := NewRequester(RequesterOptions{
		Client:           c,
		TargetWorkflowID: "some-workflow",
	})
	require.NoError(t, err)
	defer req.Close()

	// Request
	res, err := req.RequestUppercase(context.Background(), "SoMe VaLuE")
	require.NoError(t, err)
	require.Equal(t, "SOME VALUE", res)
}
