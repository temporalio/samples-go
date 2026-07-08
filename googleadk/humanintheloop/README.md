### Google ADK human-in-the-loop (durable approval)

A **human-in-the-loop** Google ADK agent running durably on Temporal with the
[`googleadk`](https://pkg.go.dev/go.temporal.io/sdk/contrib/googleadk) contrib
integration. The agent has a sensitive `delete_resource` function tool that must
not run without a human's approval.

How it works:

1. The model calls `delete_resource`. On its **first** invocation the tool sees
   `ctx.ToolConfirmation() == nil`, calls `ctx.RequestConfirmation("Delete …?", nil)`,
   and returns without doing the delete — so ADK pauses the agent.
2. `ApprovalWorkflow` detects the pause via `googleadk.PendingConfirmations` and
   **durably blocks on a Temporal signal** (`googleadk.ConfirmationSignalName`,
   read with `workflow.GetSignalChannel`) carrying a
   `googleadk.ConfirmationDecision`.
3. When the decision arrives, the workflow resumes the run with
   `googleadk.ConfirmationResponse(decision)`. ADK re-dispatches the original
   tool call, which now sees a confirmation and performs the delete (or is blocked
   if denied).

**This is the differentiator: the wait for the human is durable.** The workflow
can sit blocked for minutes or days and survive worker restarts — no state is lost.
When the approval signal finally arrives, the agent resumes exactly where it paused.

### Notes

- **`delete_resource` runs in-workflow and only *simulates* the delete** (it returns
  a status map) to keep the demo deterministic. A real destructive operation does I/O
  and must not run in the workflow — expose it with `googleadk.ActivityAsTool` so it
  runs worker-side under Temporal's retry/timeout policy. The confirmation gate is
  identical either way: the tool still calls `ctx.RequestConfirmation(...)` before
  doing the work.
- **The workflow handles one pending confirmation per resume pass.** If a single
  model turn asked to approve several tool calls at once, you would collect a
  decision for each and pass them together to
  `googleadk.ConfirmationResponse(decisions...)`. This sample keeps to the common
  single-confirmation case.

### Prerequisites

- A running [Temporal server](https://github.com/temporalio/samples-go/tree/main/#how-to-use)
  (e.g. `temporal server start-dev`).
- A Gemini API key from <https://aistudio.google.com/apikey>, exported worker-side.

### Steps to run this sample

1) Start a Temporal server (see prerequisites).

2) In a second terminal, start the worker:
```bash
export GEMINI_API_KEY=...
go run googleadk/humanintheloop/worker/main.go
```

3) In a third terminal, run the starter:
```bash
go run googleadk/humanintheloop/starter/main.go
```

The starter asks the agent to delete a resource; the workflow pauses awaiting
approval, and the starter then sends an approval signal (via
`client.SignalWorkflow`) to demonstrate the resume. In a real system the signal
would come from an operator clicking "approve" in a UI, possibly much later.

### Test without a live LLM

`workflow_test.go` scripts the model to call `delete_resource`, uses
`env.RegisterDelayedCallback` to deliver the approval through the **real** Temporal
signal, and asserts the delete completes only after approval (plus a denial case).
No API key or network needed:
```bash
go test ./googleadk/humanintheloop/...
```
