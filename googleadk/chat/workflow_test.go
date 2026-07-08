package chat_test

import (
	"context"
	"iter"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"

	"google.golang.org/adk/v2/model"

	"go.temporal.io/sdk/contrib/googleadk"

	chat "github.com/temporalio/samples-go/googleadk/chat"
)

// recordingModel wraps a FakeModel and records the number of Contents in each
// request it serves, so a test can prove that a later turn's request carried the
// prior conversation history (proving the session persisted across signals).
type recordingModel struct {
	inner *googleadk.FakeModel

	mu              sync.Mutex
	requestContents []int
}

func (m *recordingModel) Name() string { return "recording-model" }

func (m *recordingModel) GenerateContent(ctx context.Context, req *model.LLMRequest, stream bool) iter.Seq2[*model.LLMResponse, error] {
	if req != nil {
		m.mu.Lock()
		m.requestContents = append(m.requestContents, len(req.Contents))
		m.mu.Unlock()
	}
	return m.inner.GenerateContent(ctx, req, stream)
}

func (m *recordingModel) contentsAt(turn int) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	if turn < 0 || turn >= len(m.requestContents) {
		return -1
	}
	return m.requestContents[turn]
}

// TestChatCarriesHistory drives two user messages through the chat workflow via
// signals and asserts the second turn's model request carried more Contents than
// the first — proving the two messages ran on the SAME ADK session, so history
// accumulated. MaxTurns is high enough that no continue-as-new fires here.
func TestChatCarriesHistory(t *testing.T) {
	var s testsuite.WorkflowTestSuite
	env := s.NewTestWorkflowEnvironment()
	env.RegisterWorkflow(chat.ChatWorkflow)

	rec := &recordingModel{inner: googleadk.NewFakeModel(
		googleadk.TextResponse("Hi David, nice to meet you!"),
		googleadk.TextResponse("Durable execution means your program's state survives crashes."),
	)}
	acts, err := googleadk.NewActivities(googleadk.Config{
		Models: map[string]googleadk.ModelFactory{
			chat.ModelName: func(context.Context, string) (model.LLM, error) { return rec, nil },
		},
	})
	require.NoError(t, err)
	env.RegisterActivityWithOptions(acts.InvokeModel, activity.RegisterOptions{Name: googleadk.InvokeModelActivityName})

	// Send two messages, spaced so each is served before the next arrives. Use a
	// high MaxTurns so this test does not hit the continue-as-new path.
	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow(chat.UserMessageSignalName, "Hi! My name is David.")
	}, time.Second)
	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow(chat.UserMessageSignalName, "What is durable execution?")
	}, 5*time.Second)
	// End the (otherwise infinite) workflow by cancelling after both are served.
	env.RegisterDelayedCallback(func() {
		env.CancelWorkflow()
	}, 10*time.Second)

	env.ExecuteWorkflow(chat.ChatWorkflow, chat.ChatInput{MaxTurns: 100})

	require.True(t, env.IsWorkflowCompleted())

	// Turn 1's request had just the first user message. Turn 2's request also
	// carried turn 1's user message + the model's reply — strictly more Contents.
	first := rec.contentsAt(0)
	second := rec.contentsAt(1)
	require.GreaterOrEqual(t, first, 1, "the first turn must have served at least one request")
	require.GreaterOrEqual(t, second, 1, "the second turn must have served at least one request")
	assert.Greater(t, second, first, "the second turn's request must carry prior history (same session)")
}

// TestChatContinueAsNew exercises the bounded-history path: with MaxTurns=1 the
// workflow serves one message and then continues-as-new, exporting the session. In
// the test environment a continue-as-new surfaces as a workflow error of type
// *ContinueAsNewError, which is the assertion here.
func TestChatContinueAsNew(t *testing.T) {
	var s testsuite.WorkflowTestSuite
	env := s.NewTestWorkflowEnvironment()
	env.RegisterWorkflow(chat.ChatWorkflow)

	fm := googleadk.NewFakeModel(googleadk.TextResponse("Hello!"))
	acts, err := googleadk.NewActivities(googleadk.Config{
		Models: map[string]googleadk.ModelFactory{
			chat.ModelName: func(context.Context, string) (model.LLM, error) { return fm, nil },
		},
	})
	require.NoError(t, err)
	env.RegisterActivityWithOptions(acts.InvokeModel, activity.RegisterOptions{Name: googleadk.InvokeModelActivityName})

	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow(chat.UserMessageSignalName, "Hi there!")
	}, time.Second)

	env.ExecuteWorkflow(chat.ChatWorkflow, chat.ChatInput{MaxTurns: 1})

	require.True(t, env.IsWorkflowCompleted())
	// After serving the single allowed turn, the workflow must continue-as-new.
	err = env.GetWorkflowError()
	require.Error(t, err, "MaxTurns=1 must trigger a continue-as-new")
	var canErr *workflow.ContinueAsNewError
	require.ErrorAs(t, err, &canErr, "the workflow must end by continuing-as-new to bound history")
}
