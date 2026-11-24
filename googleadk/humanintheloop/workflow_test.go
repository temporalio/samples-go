package humanintheloop_test

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

	"google.golang.org/adk/v2/model"
	"google.golang.org/genai"

	"go.temporal.io/googleadk"

	humanintheloop "github.com/temporalio/samples-go/googleadk/humanintheloop"
)

// scriptedModelFactory returns a ModelFactory yielding a single shared FakeModel
// so its scripted responses advance turn by turn across Activity invocations.
func scriptedModelFactory(responses ...*model.LLMResponse) googleadk.ModelFactory {
	fm := googleadk.NewFakeModel(responses...)
	return func(context.Context, string) (model.LLM, error) { return fm, nil }
}

type recordingModel struct {
	*googleadk.FakeModel
	mu       sync.Mutex
	requests []*model.LLMRequest
}

func (m *recordingModel) GenerateContent(ctx context.Context, req *model.LLMRequest, stream bool) iter.Seq2[*model.LLMResponse, error] {
	m.mu.Lock()
	m.requests = append(m.requests, req)
	m.mu.Unlock()
	return m.FakeModel.GenerateContent(ctx, req, stream)
}

// TestApprovalWorkflow proves the durable human-in-the-loop wait: the agent calls
// the sensitive delete_resource tool, which pauses the workflow awaiting a human
// decision. A delayed callback delivers an approval via the real Temporal signal
// (googleadk.ConfirmationSignalName), after which the workflow resumes and the
// delete runs. This exercises the actual signal round-trip, not a two-pass loop.
func TestApprovalWorkflow(t *testing.T) {
	var s testsuite.WorkflowTestSuite
	env := s.NewTestWorkflowEnvironment()

	env.RegisterWorkflow(humanintheloop.ApprovalWorkflow)

	// Scripted model: turn 1 (before approval) calls delete_resource; turn 2
	// (after the resume) produces the final text confirming the delete.
	acts, err := googleadk.NewActivities(googleadk.Config{
		Models: map[string]googleadk.ModelFactory{
			humanintheloop.ModelName: scriptedModelFactory(
				googleadk.FunctionCallResponse("call-1", humanintheloop.DeleteToolName, map[string]any{"resource": "prod-db"}),
				googleadk.TextResponse("Deleted prod-db."),
			),
		},
	})
	require.NoError(t, err)
	env.RegisterActivityWithOptions(acts.InvokeModel, activity.RegisterOptions{Name: googleadk.InvokeModelActivityName})

	// Approve after a delay, via the real signal the workflow blocks on. Until this
	// fires the workflow is durably waiting; the delete cannot have run yet.
	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow(googleadk.ConfirmationSignalName, googleadk.ConfirmationDecision{Confirmed: true})
	}, time.Second)

	env.ExecuteWorkflow(humanintheloop.ApprovalWorkflow, "Please delete the resource named prod-db.")

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())

	var res humanintheloop.Result
	require.NoError(t, env.GetWorkflowResult(&res))
	assert.True(t, res.Approved, "the workflow must record the human's approval")
	// The final answer is produced only on the resumed pass — after approval.
	assert.Contains(t, res.Answer, "Deleted")
}

// TestApprovalWorkflowDenied proves denial also flows through: the human denies,
// the workflow records it, and the delete does not report success.
func TestApprovalWorkflowDenied(t *testing.T) {
	var s testsuite.WorkflowTestSuite
	env := s.NewTestWorkflowEnvironment()

	env.RegisterWorkflow(humanintheloop.ApprovalWorkflow)

	acts, err := googleadk.NewActivities(googleadk.Config{
		Models: map[string]googleadk.ModelFactory{
			humanintheloop.ModelName: scriptedModelFactory(
				googleadk.FunctionCallResponse("call-1", humanintheloop.DeleteToolName, map[string]any{"resource": "prod-db"}),
				googleadk.TextResponse("Okay, I did not delete prod-db."),
			),
		},
	})
	require.NoError(t, err)
	env.RegisterActivityWithOptions(acts.InvokeModel, activity.RegisterOptions{Name: googleadk.InvokeModelActivityName})

	// A dropped signal must not leave fields behind when the next JSON signal
	// omits them. In particular, the denial must not inherit the stale ID.
	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow(googleadk.ConfirmationSignalName, map[string]any{"FunctionCallID": "stale", "Confirmed": true})
	}, time.Second)
	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow(googleadk.ConfirmationSignalName, map[string]any{"Confirmed": false})
	}, 2*time.Second)

	env.ExecuteWorkflow(humanintheloop.ApprovalWorkflow, "Please delete the resource named prod-db.")

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())

	var res humanintheloop.Result
	require.NoError(t, env.GetWorkflowResult(&res))
	assert.False(t, res.Approved, "the workflow must record the human's denial")
}

// TestApprovalWorkflowTwoConfirmations checks that the workflow waits for both
// decisions before resuming, and that each tool's actual result reaches the model.
func TestApprovalWorkflowTwoConfirmations(t *testing.T) {
	for _, tc := range []struct {
		name         string
		prodApproved bool
	}{
		{name: "both approved", prodApproved: true},
		{name: "denied then approved", prodApproved: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var s testsuite.WorkflowTestSuite
			env := s.NewTestWorkflowEnvironment()
			env.RegisterWorkflow(humanintheloop.ApprovalWorkflow)

			twoDeletes := googleadk.FunctionCallResponse("call-1", humanintheloop.DeleteToolName, map[string]any{"resource": "prod-db"})
			twoDeletes.Content.Parts = append(twoDeletes.Content.Parts, &genai.Part{
				FunctionCall: &genai.FunctionCall{ID: "call-2", Name: humanintheloop.DeleteToolName, Args: map[string]any{"resource": "staging-db"}},
			})
			fm := &recordingModel{FakeModel: googleadk.NewFakeModel(twoDeletes, googleadk.TextResponse("Decisions applied."))}
			acts, err := googleadk.NewActivities(googleadk.Config{
				Models: map[string]googleadk.ModelFactory{
					humanintheloop.ModelName: func(context.Context, string) (model.LLM, error) { return fm, nil },
				},
			})
			require.NoError(t, err)
			env.RegisterActivityWithOptions(acts.InvokeModel, activity.RegisterOptions{Name: googleadk.InvokeModelActivityName})

			pending := func() []googleadk.PendingConfirmation {
				value, err := env.QueryWorkflow(humanintheloop.PendingConfirmationsQuery)
				require.NoError(t, err)
				var requests []googleadk.PendingConfirmation
				require.NoError(t, value.Get(&requests))
				return requests
			}
			var firstID string
			env.RegisterDelayedCallback(func() {
				requests := pending()
				require.Len(t, requests, 2)
				assert.Equal(t, "prod-db", requests[0].OriginalCall.Args["resource"])
				assert.Equal(t, "Delete prod-db?", requests[0].Hint)
				// This signal is ambiguous and must leave both requests pending.
				env.SignalWorkflow(googleadk.ConfirmationSignalName, googleadk.ConfirmationDecision{Confirmed: true})
			}, time.Second)
			env.RegisterDelayedCallback(func() {
				requests := pending()
				require.Len(t, requests, 2)
				firstID = requests[0].FunctionCallID
				env.SignalWorkflow(googleadk.ConfirmationSignalName, googleadk.ConfirmationDecision{FunctionCallID: firstID, Confirmed: tc.prodApproved})
			}, 2*time.Second)
			env.RegisterDelayedCallback(func() {
				requests := pending()
				require.Len(t, requests, 1)
				assert.Equal(t, "staging-db", requests[0].OriginalCall.Args["resource"])
				// A duplicate must neither answer staging-db nor change the first decision.
				env.SignalWorkflow(googleadk.ConfirmationSignalName, googleadk.ConfirmationDecision{FunctionCallID: firstID, Confirmed: !tc.prodApproved})
			}, 3*time.Second)
			env.RegisterDelayedCallback(func() {
				require.Len(t, pending(), 1)
				// With only one request left, omitting the ID is unambiguous.
				env.SignalWorkflow(googleadk.ConfirmationSignalName, map[string]any{"Confirmed": true})
			}, 4*time.Second)

			env.ExecuteWorkflow(humanintheloop.ApprovalWorkflow, "Please delete prod-db and staging-db.")
			require.True(t, env.IsWorkflowCompleted())
			require.NoError(t, env.GetWorkflowError())
			assert.Empty(t, pending())

			var res humanintheloop.Result
			require.NoError(t, env.GetWorkflowResult(&res))
			assert.Equal(t, tc.prodApproved, res.Approved)
			assert.Equal(t, "Decisions applied.", res.Answer)

			fm.mu.Lock()
			defer fm.mu.Unlock()
			require.Len(t, fm.requests, 2, "the model must resume only after both decisions")
			deleted := map[string]string{}
			for _, content := range fm.requests[1].Contents {
				for _, part := range content.Parts {
					if fr := part.FunctionResponse; fr != nil && fr.Name == humanintheloop.DeleteToolName && fr.Response["status"] == "deleted" {
						deleted[fr.ID] = fr.Response["resource"].(string)
					}
				}
			}
			wantDeleted := map[string]string{"call-2": "staging-db"}
			if tc.prodApproved {
				wantDeleted["call-1"] = "prod-db"
			}
			assert.Equal(t, wantDeleted, deleted, "tool results must match the approvals")
		})
	}
}
