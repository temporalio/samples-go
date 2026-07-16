package sleepfordays

import (
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"go.temporal.io/sdk/testsuite"
)

func TestSleepForDaysWorkflow(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	numActivityCalls := 0
	env.RegisterActivity(SendEmailActivity)
	env.OnActivity(SendEmailActivity, mock.Anything, mock.Anything).Run(
		func(args mock.Arguments) { numActivityCalls++ },
	).Return(nil)

	startTime := env.Now()

	// Time-skip 90 days.
	env.RegisterDelayedCallback(func() {
		// Check that the activity has been called 3 times.
		require.Equal(t, 3, numActivityCalls)
		// Send the signal to complete the workflow.
		env.SignalWorkflow("complete", nil)
		// Expect no more activity calls to have been made - workflow is complete.
		require.Equal(t, 3, numActivityCalls)
		// Expect more than 90 days to have passed.
		require.Equal(t, env.Now().Sub(startTime), time.Hour*24*90)
	}, time.Hour*24*90)

	// Execute workflow.
	env.ExecuteWorkflow(SleepForDaysWorkflow)
}
