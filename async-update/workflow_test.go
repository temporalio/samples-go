package async_update_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	async_update "github.com/temporalio/samples-go/async-update"
	"go.temporal.io/sdk/testsuite"
)

type updateCallback struct {
	accept   func()
	reject   func(error)
	complete func(interface{}, error)
}

func (uc *updateCallback) Accept() {
	uc.accept()
}

func (uc *updateCallback) Reject(err error) {
	uc.reject(err)
}

func (uc *updateCallback) Complete(success interface{}, err error) {
	uc.complete(success, err)
}

func TestWorkflow(t *testing.T) {
	// Create env
	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestWorkflowEnvironment()
	env.RegisterWorkflow(async_update.ProcessWorkflow)
	env.OnActivity(async_update.Activity, mock.Anything, "world").After(5*time.Second).Return("hello world", nil)

	// Use delayed callbacks to send multiple updates at the same time
	for i := 0; i < 10; i++ {
		i := i
		env.RegisterDelayedCallback(func() {
			env.UpdateWorkflow(async_update.ProcessUpdateName, "test id", &updateCallback{
				accept: func() {
					if i >= 5 {
						require.Fail(t, "update should fail since we should exceed our max update limit")
					}
				},
				reject: func(err error) {
					if i < 5 {
						require.Fail(t, "this update should not fail")
					}
					require.Error(t, err)
				},
				complete: func(response interface{}, err error) {
					require.NoError(t, err)
					require.Equal(t, "hello world", response)
				},
			}, "world")
		}, 0)
	}
	// Use delayed callback to send signal
	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow(async_update.Done, nil)
	}, time.Second)

	// Send an update after the workflow is signaled to close, expect the update to be rejected
	env.RegisterDelayedCallback(func() {
		env.UpdateWorkflow(async_update.ProcessUpdateName, "test id", &updateCallback{
			accept: func() {
				require.Fail(t, "update should fail since the workflow is closing")
			},
			reject: func(err error) {
				require.Error(t, err)
			},
			complete: func(response interface{}, err error) {
			},
		}, "world")
	}, 2*time.Second)

	// Run workflow
	env.ExecuteWorkflow(async_update.ProcessWorkflow)
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	var result int
	require.NoError(t, env.GetWorkflowResult(&result))
	require.Equal(t, 5, result)
}
