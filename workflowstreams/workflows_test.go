package streams

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/contrib/workflowstreams"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/worker"
)

// The unit tests below cover the workflow side only. The client Subscribe path
// needs a live Temporal client, so it is exercised by Test_OrderWorkflow_DevServer.

func Test_OrderWorkflow(t *testing.T) {
	env := (&testsuite.WorkflowTestSuite{}).NewTestWorkflowEnvironment()
	env.OnActivity(ChargeCard, mock.Anything, "order-42").Return("charge-order-42", nil)

	env.ExecuteWorkflow(OrderWorkflow, OrderInput{OrderID: "order-42"})

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	var result string
	require.NoError(t, env.GetWorkflowResult(&result))
	require.Equal(t, "charge-order-42", result)
}

func Test_PipelineWorkflow(t *testing.T) {
	env := (&testsuite.WorkflowTestSuite{}).NewTestWorkflowEnvironment()

	env.ExecuteWorkflow(PipelineWorkflow, PipelineInput{PipelineID: "p1"})

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	var result string
	require.NoError(t, env.GetWorkflowResult(&result))
	require.Equal(t, "pipeline p1 done", result)
}

func Test_HubWorkflow(t *testing.T) {
	env := (&testsuite.WorkflowTestSuite{}).NewTestWorkflowEnvironment()

	// Signal the workflow to close after a simulated delay.
	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow(CloseSignal, nil)
	}, time.Second)

	env.ExecuteWorkflow(HubWorkflow, HubInput{HubID: "newsroom"})

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	var result string
	require.NoError(t, env.GetWorkflowResult(&result))
	require.Equal(t, "hub newsroom closed", result)
}

func Test_TickerWorkflow(t *testing.T) {
	env := (&testsuite.WorkflowTestSuite{}).NewTestWorkflowEnvironment()

	env.ExecuteWorkflow(TickerWorkflow, TickerInput{Count: 12, KeepLast: 5, TruncateEvery: 5})

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	var result string
	require.NoError(t, env.GetWorkflowResult(&result))
	require.Equal(t, "ticker emitted 12 events", result)
}

func Test_LLMWorkflow(t *testing.T) {
	env := (&testsuite.WorkflowTestSuite{}).NewTestWorkflowEnvironment()
	env.OnActivity(StreamCompletion, mock.Anything, mock.Anything).Return("a streamed answer", nil)

	env.ExecuteWorkflow(LLMWorkflow, LLMInput{Prompt: "hello"})

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	var result string
	require.NoError(t, env.GetWorkflowResult(&result))
	require.Equal(t, "a streamed answer", result)
}

// Test_OrderWorkflow_DevServer runs the basic publish/subscribe scenario end to
// end against a local dev server: it exercises publishing from both the workflow
// and the activity (via the real client) and consuming via Subscribe.
func Test_OrderWorkflow_DevServer(t *testing.T) {
	server, err := testsuite.StartDevServer(context.Background(), testsuite.DevServerOptions{
		ClientOptions: &client.Options{HostPort: ""},
	})
	require.NoError(t, err)
	defer func() { _ = server.Stop() }()

	c := server.Client()
	w := worker.New(c, TaskQueue, worker.Options{})
	w.RegisterWorkflow(OrderWorkflow)
	w.RegisterActivity(ChargeCard)
	require.NoError(t, w.Start())
	defer w.Stop()

	ctx := context.Background()
	workflowID := "workflow-streams-order-test"
	_, err = c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: TaskQueue,
	}, OrderWorkflow, OrderInput{OrderID: "order-42"})
	require.NoError(t, err)

	dc := converter.GetDefaultDataConverter()
	stream := workflowstreams.NewClient(c, workflowID, workflowstreams.Options{})
	defer func() { _ = stream.Close(ctx) }()

	var statuses []string
	progressCount := 0
	for item, err := range stream.Subscribe(ctx, workflowstreams.SubscribeOptions{
		Topics: []string{TopicStatus, TopicProgress},
	}) {
		require.NoError(t, err)
		switch item.Topic {
		case TopicStatus:
			var evt StatusEvent
			require.NoError(t, dc.FromPayload(item.Data, &evt))
			statuses = append(statuses, evt.Kind)
			if evt.Kind == "complete" {
				require.Equal(t, []string{"received", "shipped", "complete"}, statuses)
				require.GreaterOrEqual(t, progressCount, 2)
				return
			}
		case TopicProgress:
			var evt ProgressEvent
			require.NoError(t, dc.FromPayload(item.Data, &evt))
			progressCount++
		}
	}
}
