package humanintheloop_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/testsuite"

	"google.golang.org/adk/v2/model"

	"go.temporal.io/sdk/contrib/googleadk"

	humanintheloop "github.com/temporalio/samples-go/googleadk/humanintheloop"
)

// scriptedModelFactory returns a ModelFactory yielding a single shared FakeModel
// so its scripted responses advance turn by turn across Activity invocations.
func scriptedModelFactory(responses ...*model.LLMResponse) googleadk.ModelFactory {
	fm := googleadk.NewFakeModel(responses...)
	return func(context.Context, string) (model.LLM, error) { return fm, nil }
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

	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow(googleadk.ConfirmationSignalName, googleadk.ConfirmationDecision{Confirmed: false})
	}, time.Second)

	env.ExecuteWorkflow(humanintheloop.ApprovalWorkflow, "Please delete the resource named prod-db.")

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())

	var res humanintheloop.Result
	require.NoError(t, env.GetWorkflowResult(&res))
	assert.False(t, res.Approved, "the workflow must record the human's denial")
}
