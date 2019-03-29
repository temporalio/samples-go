package main

import (
	"github.com/stretchr/testify/require"
	"go.uber.org/cadence/testsuite"
	"testing"
	"time"
)

func Test_QueryWorkflow(t *testing.T) {
	ts := &testsuite.WorkflowTestSuite{}
	env := ts.NewTestWorkflowEnvironment()

	env.RegisterDelayedCallback(func() {
		queryAndVerify(t, env, "waiting on timer")
	}, time.Minute*5)

	env.ExecuteWorkflow(QueryWorkflow)
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	queryAndVerify(t, env, "done")
}

func queryAndVerify(t *testing.T, env *testsuite.TestWorkflowEnvironment, expectedState string) {
	result, err := env.QueryWorkflow("state")
	require.NoError(t, err)
	var state string
	err = result.Get(&state)
	require.NoError(t, err)
	require.Equal(t, expectedState, state)
}