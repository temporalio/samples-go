// Package humanintheloop demonstrates a human-in-the-loop (HITL) Google ADK
// (adk-go) agent running durably on Temporal with the
// go.temporal.io/sdk/contrib/googleadk integration. The agent has a sensitive
// delete_resource tool that pauses for human approval; the Workflow durably waits
// for a Temporal signal carrying the human's decision before letting the tool run.
//
// This is the key differentiator over a plain agent loop: the wait for the human
// is durable. The Workflow can be idle for days and survive worker restarts — when
// the approval signal finally arrives, the agent resumes exactly where it paused.
package humanintheloop

import (
	"go.temporal.io/sdk/workflow"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/agent/llmagent"
	"google.golang.org/adk/v2/runner"
	"google.golang.org/adk/v2/session"
	"google.golang.org/adk/v2/tool"
	"google.golang.org/adk/v2/tool/functiontool"
	"google.golang.org/genai"

	"go.temporal.io/sdk/contrib/googleadk"
)

const (
	// TaskQueue is the task queue the worker listens on and the starter targets.
	TaskQueue = "google-adk-hitl"

	// ModelName is the Gemini model name the agent ships in-workflow.
	ModelName = "gemini-2.0-flash"

	// DeleteToolName is the name of the sensitive function tool the model calls.
	DeleteToolName = "delete_resource"
)

// DeleteArgs is the argument schema the model fills in for the delete tool.
type DeleteArgs struct {
	Resource string `json:"resource"`
}

// Result is the serializable output of ApprovalWorkflow.
type Result struct {
	// Approved reports the human's decision.
	Approved bool
	// Answer is the agent's final text after the decision was applied.
	Answer string
}

// deleteResource is the sensitive function tool. On its first invocation there is
// no confirmation yet, so it requests one (via ctx.RequestConfirmation) and
// returns without doing the work — this pauses the agent. On the resumed
// invocation ADK supplies a ToolConfirmation, so the delete proceeds.
func deleteResource(tctx agent.Context, args DeleteArgs) (map[string]any, error) {
	if tctx.ToolConfirmation() == nil {
		if err := tctx.RequestConfirmation("Delete "+args.Resource+"?", nil); err != nil {
			return nil, err
		}
		return map[string]any{"status": "awaiting confirmation"}, nil
	}
	return map[string]any{"status": "deleted", "resource": args.Resource}, nil
}

// ApprovalWorkflow runs the agent and, when the sensitive tool pauses awaiting a
// human decision, durably blocks on a Temporal signal named
// googleadk.ConfirmationSignalName carrying a googleadk.ConfirmationDecision. Once
// the decision arrives it resumes the agent with googleadk.ConfirmationResponse,
// so the tool runs (or is blocked) according to the human's choice.
func ApprovalWorkflow(ctx workflow.Context, request string) (Result, error) {
	delTool, err := functiontool.New[DeleteArgs, map[string]any](
		functiontool.Config{
			Name:        DeleteToolName,
			Description: "Delete a named resource. Requires human confirmation before it runs.",
		},
		deleteResource,
	)
	if err != nil {
		return Result{}, err
	}

	root, err := llmagent.New(llmagent.Config{
		Name:        "assistant",
		Description: "an assistant that can delete resources with human approval",
		Model:       googleadk.NewModel(ModelName),
		Instruction: "Use the delete_resource tool when the user asks to delete something.",
		Tools:       []tool.Tool{delTool},
	})
	if err != nil {
		return Result{}, err
	}

	r, err := runner.New(runner.Config{
		AppName:           "hitl",
		Agent:             root,
		SessionService:    session.InMemoryService(),
		AutoCreateSession: true,
	})
	if err != nil {
		return Result{}, err
	}

	adkCtx := googleadk.NewContext(ctx)
	msg := genai.NewContentFromText(request, genai.RoleUser)

	var res Result
	// Drive the run in passes: each Run call is one pass over the same session. A
	// pass either completes (no pending confirmation) or pauses awaiting a human.
	for {
		var events []*session.Event
		for ev, err := range r.Run(adkCtx, "user-1", "session-1", msg, agent.RunConfig{}) {
			if err != nil {
				return Result{}, err
			}
			if ev == nil {
				continue
			}
			events = append(events, ev)
			if ev.Content != nil {
				for _, p := range ev.Content.Parts {
					if p != nil && p.Text != "" {
						res.Answer = p.Text
					}
				}
			}
		}

		pending := googleadk.PendingConfirmations(events)
		if len(pending) == 0 {
			// The agent finished without (further) confirmations needed.
			return res, nil
		}

		// The agent paused. Durably wait for the human's decision to arrive as a
		// Temporal signal. This is the whole point: the workflow can sit here for
		// as long as it takes — across worker restarts — without losing state.
		var decision googleadk.ConfirmationDecision
		workflow.GetSignalChannel(ctx, googleadk.ConfirmationSignalName).Receive(ctx, &decision)
		res.Approved = decision.Confirmed

		// Match the decision to the pending confirmation and resume the run with
		// it as the next message. ADK re-dispatches (or blocks) the original tool
		// call based on Confirmed.
		if decision.FunctionCallID == "" {
			decision.FunctionCallID = pending[0].FunctionCallID
		}
		msg = googleadk.ConfirmationResponse(decision)
	}
}
