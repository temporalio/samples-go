// Package chat demonstrates a long-lived, update-driven Google ADK (adk-go) chat
// running durably on Temporal with the go.temporal.io/sdk/contrib/googleadk
// integration. A single Workflow serves an ongoing conversation: each user message
// arrives as a Temporal Update, the agent answers it on the SAME ADK session (so
// history accumulates), and the answer is returned on the Update itself — no signal
// + query polling.
//
// To keep history bounded, the Workflow continues-as-new once Temporal suggests it
// (or after a demo turn cap): it exports the session with googleadk.ExportSession,
// then re-imports it at the top of the next run with googleadk.ImportSession — so
// the conversation carries across the continue-as-new boundary without the history
// growing unbounded in a single run.
package chat

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/agent/llmagent"
	"google.golang.org/adk/v2/runner"
	"google.golang.org/adk/v2/session"
	"google.golang.org/genai"

	"go.temporal.io/sdk/contrib/googleadk"
)

const (
	// TaskQueue is the task queue the worker listens on and the starter targets.
	TaskQueue = "google-adk-chat"

	// ModelName is the Gemini model name the agent ships in-workflow.
	ModelName = "gemini-2.0-flash"

	// SendMessageUpdateName is the Update that delivers a user message and returns
	// the agent's answer. An Update (rather than a signal + query) lets the caller
	// send the message and receive the answer on one call, with no polling.
	SendMessageUpdateName = "send-message"

	// AppName / UserID / SessionID identify the single conversation session.
	AppName   = "chat"
	UserID    = "user-1"
	SessionID = "session-1"
)

// ChatInput is the workflow argument. On first start Snapshot is nil; on a
// continue-as-new it carries the exported session so the conversation resumes.
type ChatInput struct {
	// Snapshot, when non-nil, is the session exported by the previous run.
	Snapshot *googleadk.SessionSnapshot
	// MaxTurns caps the number of messages served before continuing-as-new, so
	// the demo can force the boundary without waiting for Temporal's suggestion.
	// Zero means "only continue-as-new when Temporal suggests it".
	MaxTurns int
}

// ChatWorkflow serves a long-lived conversation. It imports any prior session, then
// registers a SendMessage Update handler that runs one agent turn per message on the
// shared session and returns the answer. When Temporal suggests continue-as-new (or
// MaxTurns is reached) it drains any in-flight turn, exports the session, and
// continues-as-new carrying the snapshot forward.
// @@@SNIPSTART googleadk-chat-workflow
func ChatWorkflow(ctx workflow.Context, in ChatInput) error {
	// A fresh in-memory session service, kept in a local so we can Export it later.
	svc := session.InMemoryService()

	adkCtx := googleadk.NewContext(ctx)

	// Resume a prior conversation if this run was continued-as-new.
	if in.Snapshot != nil {
		if _, err := googleadk.ImportSession(adkCtx, svc, in.Snapshot); err != nil {
			return err
		}
	}

	root, err := llmagent.New(llmagent.Config{
		Name:        "assistant",
		Description: "a friendly conversational assistant",
		Model:       googleadk.NewModel(ModelName),
		Instruction: "You are a helpful assistant. Answer the user, using the conversation history for context.",
	})
	if err != nil {
		return err
	}

	r, err := runner.New(runner.Config{
		AppName:           AppName,
		Agent:             root,
		SessionService:    svc,
		AutoCreateSession: true,
	})
	if err != nil {
		return err
	}

	turns := 0
	// One agent turn runs at a time: serialize concurrent Updates so they can't
	// interleave on the shared ADK session.
	busy := false

	err = workflow.SetUpdateHandlerWithOptions(
		ctx,
		SendMessageUpdateName,
		func(ctx workflow.Context, text string) (string, error) {
			if err := workflow.Await(ctx, func() bool { return !busy }); err != nil {
				return "", err
			}
			busy = true
			defer func() { busy = false }()

			// Build the ADK context from this Update handler's own workflow.Context so
			// the model Activity is scheduled on the handler's coroutine.
			turnCtx := googleadk.NewContext(ctx)
			var answer string
			msg := genai.NewContentFromText(text, genai.RoleUser)
			for ev, err := range r.Run(turnCtx, UserID, SessionID, msg, agent.RunConfig{}) {
				if err != nil {
					return "", err
				}
				if ev == nil || ev.Content == nil {
					continue
				}
				for _, p := range ev.Content.Parts {
					if p != nil && p.Text != "" {
						answer = p.Text
					}
				}
			}
			turns++
			return answer, nil
		},
		workflow.UpdateHandlerOptions{
			Validator: func(ctx workflow.Context, text string) error {
				if text == "" {
					return fmt.Errorf("message must not be empty")
				}
				return nil
			},
		},
	)
	if err != nil {
		return err
	}

	// Serve messages until Temporal suggests continue-as-new (history getting large)
	// or the demo turn cap is reached.
	if err := workflow.Await(ctx, func() bool {
		return workflow.GetInfo(ctx).GetContinueAsNewSuggested() || (in.MaxTurns > 0 && turns >= in.MaxTurns)
	}); err != nil {
		return err
	}

	// Let any in-flight Update finish so its turn is captured in the snapshot.
	if err := workflow.Await(ctx, func() bool { return workflow.AllHandlersFinished(ctx) }); err != nil {
		return err
	}

	snap, err := googleadk.ExportSession(adkCtx, svc, AppName, UserID, SessionID)
	if err != nil {
		return err
	}
	return workflow.NewContinueAsNewError(ctx, ChatWorkflow, ChatInput{
		Snapshot: snap,
		MaxTurns: in.MaxTurns,
	})
}

// @@@SNIPEND
