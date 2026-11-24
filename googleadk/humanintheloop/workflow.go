// Package humanintheloop demonstrates a human-in-the-loop (HITL) Google ADK
// (adk-go) agent running durably on Temporal with the
// go.temporal.io/googleadk integration. The agent has a sensitive
// delete_resource tool that pauses for human approval; the Workflow durably waits
// for a Temporal signal carrying the human's decision before letting the tool run.
//
// This is the key differentiator over a plain agent loop: the wait for the human
// is durable. The Workflow can be idle for days and survive worker restarts — when
// the approval signal finally arrives, the agent resumes exactly where it paused.
package humanintheloop

import (
	"slices"

	"go.temporal.io/sdk/workflow"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/agent/llmagent"
	"google.golang.org/adk/v2/runner"
	"google.golang.org/adk/v2/session"
	"google.golang.org/adk/v2/tool"
	"google.golang.org/adk/v2/tool/functiontool"
	"google.golang.org/genai"

	"go.temporal.io/googleadk"
)

const (
	// TaskQueue is the task queue the worker listens on and the starter targets.
	TaskQueue = "google-adk-hitl"

	// ModelName is the Gemini model name the agent ships in-workflow.
	ModelName = "gemini-3.8-flash"

	// DeleteToolName is the name of the sensitive function tool the model calls.
	DeleteToolName = "delete_resource"

	// PendingConfirmationsQuery returns the requests still awaiting a decision.
	PendingConfirmationsQuery = "pending-confirmations"
)

// DeleteArgs is the argument schema the model fills in for the delete tool.
type DeleteArgs struct {
	Resource string `json:"resource"`
}

// Result is the serializable output of ApprovalWorkflow.
type Result struct {
	// Approved is true when at least one decision was received and all were approvals.
	Approved bool
	// Answer is the agent's final text after the decision was applied.
	Answer string
}

// deleteResource is the sensitive function tool. On its first invocation there is
// no confirmation yet, so it requests one (via ctx.RequestConfirmation) and
// returns without doing the work — this pauses the agent. On the resumed
// invocation ADK supplies a ToolConfirmation, so the delete proceeds.
//
// NOTE: this runs in-workflow and only simulates the delete (it returns a status
// map), which keeps the sample deterministic. A real destructive operation does
// I/O and must NOT run in the workflow — expose it with googleadk.ActivityAsTool
// so it runs worker-side under Temporal's retry/timeout policy. The confirmation
// gate is identical either way: request confirmation first, do the work only once
// confirmed.
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
// all pending requests have decisions, it resumes the agent with
// googleadk.ConfirmationResponse, so each tool runs (or is blocked) as decided.
// @@@SNIPSTART googleadk-hitl-workflow
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
	var waiting []googleadk.PendingConfirmation
	if err := workflow.SetQueryHandler(ctx, PendingConfirmationsQuery, func() ([]googleadk.PendingConfirmation, error) {
		return waiting, nil
	}); err != nil {
		return Result{}, err
	}
	var denied bool
	decisions := workflow.GetSignalChannel(ctx, googleadk.ConfirmationSignalName)
	// Run each pass over the same session, then answer all its confirmations
	// before letting the model take another step.
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

		waiting = googleadk.PendingConfirmations(events)
		if len(waiting) == 0 {
			// The agent finished without (further) confirmations needed.
			return res, nil
		}

		// Durably wait for a decision on every pending confirmation. The workflow
		// can wait across worker restarts without losing requests or decisions.
		//
		// Omit FunctionCallID only when one request remains. Query
		// PendingConfirmationsQuery for the IDs, tool arguments, and hints.
		var batch []googleadk.ConfirmationDecision
		for len(waiting) > 0 {
			var decision googleadk.ConfirmationDecision
			decisions.Receive(ctx, &decision)
			if decision.FunctionCallID == "" {
				if len(waiting) != 1 {
					workflow.GetLogger(ctx).Warn("Ignoring confirmation without an ID: several requests are pending")
					continue
				}
				decision.FunctionCallID = waiting[0].FunctionCallID
			}
			i := slices.IndexFunc(waiting, func(p googleadk.PendingConfirmation) bool {
				return p.FunctionCallID == decision.FunctionCallID
			})
			if i < 0 {
				workflow.GetLogger(ctx).Warn("Ignoring confirmation for an unknown or already decided request", "FunctionCallID", decision.FunctionCallID)
				continue
			}
			waiting = slices.Delete(waiting, i, i+1)
			batch = append(batch, decision)
			denied = denied || !decision.Confirmed
			res.Approved = !denied
		}

		// Resume once with all decisions. ADK executes or blocks each original
		// tool call before the model takes another step.
		msg = googleadk.ConfirmationResponse(batch...)
	}
}

// @@@SNIPEND
