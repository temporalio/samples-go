## Google ADK on Temporal

Run [Google ADK](https://google.github.io/adk-docs/) (`adk-go`) agents **durably on
Temporal** using the [`googleadk`](https://pkg.go.dev/go.temporal.io/sdk/contrib/googleadk)
contrib integration. The agent's orchestration loop runs inside a Workflow; each
model call runs as a Temporal Activity (via `googleadk.NewModel`), and tools can be
ordinary Temporal activities exposed to the agent with `googleadk.ActivityAsTool`
— so model calls and tools are retried, timed-out, and visible in the Temporal UI,
and the whole run is replayable.

Agents are built the native ADK way (`llmagent.New` + `runner.New`); the only
Temporal-specific pieces are `googleadk.NewModel(...)` as the agent's model,
`googleadk.NewContext(ctx)` passed to `Run`, and the worker-side
`googleadk.NewActivities(...)` registry that holds the real Gemini client (so the
API key stays worker-side, never crossing into the workflow).

Every sample runs against a scripted `FakeModel` in its `*_test.go`, so
`go test ./googleadk/...` needs no API key or network.

### Samples

- **Basic agent** — the root files in this directory
  ([`workflow.go`](workflow.go), [`worker/`](worker), [`starter/`](starter)): a
  single agent that answers a question, calling a `get_weather` tool implemented as
  a Temporal activity via `googleadk.ActivityAsTool`.
- **[multiagent/](multiagent)** — a `coordinator` root agent that delegates to
  `weather` and `jokes` specialist SubAgents via ADK's in-workflow
  `transfer_to_agent`.
- **[humanintheloop/](humanintheloop)** — an agent with a sensitive
  `delete_resource` tool whose workflow **durably waits on a Temporal signal** for a
  human's approval before the tool runs.
- **[chat/](chat)** — a long-lived, signal-driven conversation on one ADK session
  that **continues-as-new** (exporting/importing the session) to keep history
  bounded.

### Prerequisites (for running against a live model)

- A running [Temporal server](https://github.com/temporalio/samples-go/tree/main/#how-to-use)
  (e.g. `temporal server start-dev`).
- A Gemini API key from <https://aistudio.google.com/apikey>, exported worker-side:
  ```bash
  export GEMINI_API_KEY=...
  ```

### Running the basic agent

1) Start a Temporal server (see prerequisites).

2) In a second terminal, start the worker (blocks until Ctrl+C):
```bash
export GEMINI_API_KEY=...
go run googleadk/worker/main.go
```

3) In a third terminal, run the starter:
```bash
go run googleadk/starter/main.go
```

The starter asks "What's the weather in San Francisco?"; the agent calls the
`get_weather` tool and answers. You should see a final log line similar to:
```bash
2025/12/22 15:07:25 Agent answer: It's currently sunny and about 72°F in San Francisco.
```

The exact wording comes from the model and will vary. In the Temporal UI you will
see the run's `googleadk.InvokeModel` and `get_weather` Activities.

Each scenario subdirectory has its own README with run and test instructions.

### Test without a live LLM

`workflow_test.go` (and each scenario's `*_test.go`) drives the workflow through a
scripted `FakeModel` from the `googleadk` contrib package, so it needs no API key
or network:
```bash
go test ./googleadk/...
```
