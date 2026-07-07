### Google ADK agent

Run a [Google ADK](https://google.github.io/adk-docs/) (`adk-go`) agent **durably
on Temporal** using the [`googleadk`](https://pkg.go.dev/go.temporal.io/sdk/contrib/googleadk)
contrib integration. The agent's orchestration loop runs inside a Workflow; the
model call runs as a Temporal Activity (via `googleadk.NewModel`), and the
`get_weather` tool is an ordinary Temporal activity exposed to the agent with
`googleadk.ActivityAsTool` — so both are retried, timed-out, and visible in the
Temporal UI, and the whole run is replayable.

The agent is built the native ADK way (`llmagent.New` + `runner.New`); the only
Temporal-specific pieces are `googleadk.NewModel(...)` as the agent's model,
`googleadk.NewContext(ctx)` passed to `Run`, and the worker-side
`googleadk.NewActivities(...)` registry that holds the real Gemini client.

### Prerequisites

- A running [Temporal server](https://github.com/temporalio/samples-go/tree/main/#how-to-use)
  (e.g. `temporal server start-dev`).
- A Gemini API key from <https://aistudio.google.com/apikey>, exported worker-side:
  ```bash
  export GEMINI_API_KEY=...
  ```

### Steps to run this sample

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

### Test without a live LLM

`workflow_test.go` drives the workflow through a scripted `FakeModel` (from the
`googleadk` contrib package), so it needs no API key or network:
```bash
go test ./googleadk/...
```
