package main

import (
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/cadence/testsuite"
	"testing"
	"time"
)

func Test_ProcessingWorkflow_SingleAction(t *testing.T) {
	signalData := "_1_"
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()
	// mock activityForCondition1 so it won't wait on real clock
	env.OnActivity(activityForCondition1, mock.Anything, signalData).Return("processed_1", nil)
	env.ExecuteWorkflow(ProcessingWorkflow, signalData)
	env.AssertExpectations(t)
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	var actualResult string
	require.NoError(t, env.GetWorkflowResult(&actualResult))
	require.Equal(t, "processed_1", actualResult)
}

func Test_ProcessingWorkflow_MultiAction(t *testing.T) {
	signalData := "_1_, _3_"
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()
	// mock activityForCondition1 so it won't wait on real clock
	env.OnActivity(activityForCondition1, mock.Anything, signalData).Return("processed_1", nil)
	env.OnActivity(activityForCondition3, mock.Anything, signalData).Return("processed_3", nil)
	env.ExecuteWorkflow(ProcessingWorkflow, signalData)
	env.AssertExpectations(t)
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	var actualResult string
	require.NoError(t, env.GetWorkflowResult(&actualResult))
	require.Equal(t, "processed_1processed_3", actualResult)
}

func Test_SignalHandlingWorkflow(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	env.OnActivity(activityForCondition1, mock.Anything, "_1_").Return("processed_1", nil)

	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow("trigger-signal", "_1_")
	}, time.Minute)

	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow("trigger-signal", "exit")
	}, time.Minute*2)

	env.ExecuteWorkflow(SignalHandlingWorkflow)
	env.AssertExpectations(t)
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
}
