package shoppingcart

import (
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/require"
	"testing"

	"go.temporal.io/sdk/testsuite"
)

func Test_ShoppingCartWorkflow(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()
	updatesCompleted := 0

	env.RegisterDelayedCallback(func() {
		env.UpdateWorkflow(UpdateName, uuid.New(), &testsuite.TestUpdateCallback{
			OnAccept: func() {},
			OnReject: func(err error) { panic("unexpected rejection") },
			OnComplete: func(i interface{}, err error) {
				require.NoError(t, err)
				cartState, ok := i.(CartState)
				if !ok {
					require.Fail(t, "Invalid return type")
				}
				require.Equal(t, cartState["apple"], 1)
				updatesCompleted++
			},
		}, "add", "apple")
	}, 0)

	env.RegisterDelayedCallback(func() {
		env.UpdateWorkflow(UpdateName, uuid.New(), &testsuite.TestUpdateCallback{
			OnAccept: func() {},
			OnReject: func(err error) { panic("unexpected rejection") },
			OnComplete: func(i interface{}, err error) {
				require.NoError(t, err)
				cartState, ok := i.(CartState)
				if !ok {
					require.Fail(t, "Invalid return type")
				}
				_, ok = cartState["apple"]
				require.False(t, ok)
				updatesCompleted++
			},
		}, "remove", "apple")
	}, 0)
	env.ExecuteWorkflow(CartWorkflow)

	require.True(t, env.IsWorkflowCompleted())
	require.Equal(t, updatesCompleted, 2)
}
