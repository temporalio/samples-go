### Google ADK long-lived chat (bounded history via continue-as-new)

A **long-lived, signal-driven** Google ADK chat running durably on Temporal with
the [`googleadk`](https://pkg.go.dev/go.temporal.io/sdk/contrib/googleadk) contrib
integration. A single `ChatWorkflow` serves an ongoing conversation:

- each user message arrives as a Temporal signal (`user-message`);
- the agent answers it on the **same** ADK session, so conversation history
  accumulates and later turns have full context;
- the latest answer is readable at any time via a query (`latest-answer`).

To keep a conversation from growing unbounded in one workflow run, the workflow
**continues-as-new** once Temporal suggests it
(`workflow.GetInfo(ctx).GetContinueAsNewSuggested()`) — or, for the demo, after a
small `MaxTurns` cap. Before continuing it calls `googleadk.ExportSession` to
capture the session (identity, session-scoped state, and event history) into a
serializable `googleadk.SessionSnapshot`, passes it in the `ChatInput` of the next
run, and the next run calls `googleadk.ImportSession` to rebuild the session before
serving the next message. The conversation therefore survives the boundary while
each run's history stays bounded.

### Prerequisites

- A running [Temporal server](https://github.com/temporalio/samples-go/tree/main/#how-to-use)
  (e.g. `temporal server start-dev`).
- A Gemini API key from <https://aistudio.google.com/apikey>, exported worker-side.

### Steps to run this sample

1) Start a Temporal server (see prerequisites).

2) In a second terminal, start the worker:
```bash
export GEMINI_API_KEY=...
go run googleadk/chat/worker/main.go
```

3) In a third terminal, run the starter:
```bash
go run googleadk/chat/starter/main.go
```

The starter starts the chat (with a small `MaxTurns` so continue-as-new fires
quickly), sends a couple of message signals, and queries the latest answer after
each. The workflow keeps running (continuing-as-new to bound history); terminate it
from the Temporal UI when you're done.

### Test without a live LLM

`workflow_test.go` drives two messages via signals and asserts the second turn's
model request carried prior history (proving the session persisted across signals),
and exercises the continue-as-new path with `MaxTurns=1`. No API key or network
needed:
```bash
go test ./googleadk/chat/...
```
