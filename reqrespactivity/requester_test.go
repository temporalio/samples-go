package reqrespactivity_test

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/temporalio/samples-go/reqrespactivity"
	"go.temporal.io/sdk/mocks"
	"go.temporal.io/sdk/testsuite"
)

func TestActivityRequester(t *testing.T) {
	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestActivityEnvironment()

	// Expect to be signalled
	c := &mocks.Client{}
	c.On("SignalWorkflow", mock.Anything, "some-workflow", "", "request", mock.AnythingOfType("*reqrespactivity.Request")).
		Once().
		Return(nil).
		Run(func(args mock.Arguments) {
			// Once signalled, we can now execute the activity
			req := args[len(args)-1].(*reqrespactivity.Request)
			_, err := env.ExecuteActivity(req.ResponseActivity, &reqrespactivity.Response{
				ID:     req.ID,
				Output: strings.ToUpper(req.Input),
			})
			require.NoError(t, err)
		})

	// Create requester
	req, err := reqrespactivity.NewRequester(reqrespactivity.RequesterOptions{
		Client:           c,
		TargetWorkflowID: "some-workflow",
		ExistingWorker:   &fakeWorker{env},
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
