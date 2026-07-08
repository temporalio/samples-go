// Package chat demonstrates a long-lived, signal-driven Google ADK (adk-go) chat
// running durably on Temporal with the go.temporal.io/sdk/contrib/googleadk
// integration. A single Workflow serves an ongoing conversation: each user message
// arrives as a Temporal signal, the agent answers it on the SAME ADK session (so
// history accumulates), and the latest answer is readable via a query.
//
// To keep history bounded, the Workflow continues-as-new once Temporal suggests it
// (or after a demo turn cap): it exports the session with googleadk.ExportSession,
// then re-imports it at the top of the next run with googleadk.ImportSession — so
// the conversation carries across the continue-as-new boundary without the history
// growing unbounded in a single run.
package chat

import (
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

	// UserMessageSignalName delivers a new user message to the chat workflow.
	UserMessageSignalName = "user-message"

	// LatestAnswerQueryType reads the most recent agent answer.
	LatestAnswerQueryType = "latest-answer"

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

// ChatWorkflow serves a long-lived conversation. It imports any prior session,
// then loops receiving user-message signals, running the agent per message on the
// shared session, and recording each answer (readable via a query). When Temporal
// suggests continue-as-new (or MaxTurns is reached) it exports the session and
// continues-as-new carrying the snapshot forward.
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
		AppName: AppName,
		Agent:   root,
		// AutoCreateSession creates the session on first use when we did not import
		// one; when we imported, the session already exists.
		SessionService:    svc,
		AutoCreateSession: true,
	})
	if err != nil {
		return err
	}

	// latestAnswer is served by the query handler.
	var latestAnswer string
	if err := workflow.SetQueryHandler(ctx, LatestAnswerQueryType, func() (string, error) {
		return latestAnswer, nil
	}); err != nil {
		return err
	}

	msgCh := workflow.GetSignalChannel(ctx, UserMessageSignalName)

	turns := 0
	for {
		// Wait for the next user message.
		var text string
		msgCh.Receive(ctx, &text)
		turns++

		msg := genai.NewContentFromText(text, genai.RoleUser)
		for ev, err := range r.Run(adkCtx, UserID, SessionID, msg, agent.RunConfig{}) {
			if err != nil {
				return err
			}
			if ev == nil || ev.Content == nil {
				continue
			}
			for _, p := range ev.Content.Parts {
				if p != nil && p.Text != "" {
					latestAnswer = p.Text
				}
			}
		}

		// Continue-as-new to keep history bounded: either when Temporal suggests it
		// (history/event count getting large) or after the demo turn cap.
		suggested := workflow.GetInfo(ctx).GetContinueAsNewSuggested()
		capReached := in.MaxTurns > 0 && turns >= in.MaxTurns
		if suggested || capReached {
			// Drain any messages that arrived while we were serving this turn, so
			// they aren't lost across the continue-as-new boundary.
			for {
				var pending string
				if !msgCh.ReceiveAsync(&pending) {
					break
				}
				pmsg := genai.NewContentFromText(pending, genai.RoleUser)
				for ev, err := range r.Run(adkCtx, UserID, SessionID, pmsg, agent.RunConfig{}) {
					if err != nil {
						return err
					}
					if ev == nil || ev.Content == nil {
						continue
					}
					for _, p := range ev.Content.Parts {
						if p != nil && p.Text != "" {
							latestAnswer = p.Text
						}
					}
				}
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
	}
}
