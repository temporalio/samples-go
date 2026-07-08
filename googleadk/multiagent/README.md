### Google ADK multi-agent (coordinator + specialists)

A **multi-agent** Google ADK system running durably on Temporal with the
[`googleadk`](https://pkg.go.dev/go.temporal.io/sdk/contrib/googleadk) contrib
integration. A `coordinator` root agent delegates the user's request to one of two
specialist **SubAgents**:

- a `weather` specialist that owns the `get_weather` tool (an ordinary Temporal
  activity exposed via `googleadk.ActivityAsTool`), and
- a `jokes` specialist.

The coordinator picks a specialist by emitting ADK's built-in `transfer_to_agent`
call. **That transfer is resolved in-workflow** — it is not a separate Temporal
workflow or a network round-trip — so the whole delegation is deterministic and
replayable. Only the per-agent model calls and the `get_weather` tool cross the
Activity boundary.

Each agent is given a **distinct model name** (`gemini-2.0-flash-coordinator`,
`-weather`, `-jokes`) so tests can register and script a separate `FakeModel` per
agent; in production all three names resolve to the same real Gemini model behind
`googleadk.InvokeModel`.

### Prerequisites

- A running [Temporal server](https://github.com/temporalio/samples-go/tree/main/#how-to-use)
  (e.g. `temporal server start-dev`).
- A Gemini API key from <https://aistudio.google.com/apikey>, exported worker-side.

### Steps to run this sample

1) Start a Temporal server (see prerequisites).

2) In a second terminal, start the worker:
```bash
export GEMINI_API_KEY=...
go run googleadk/multiagent/worker/main.go
```

3) In a third terminal, run the starter:
```bash
go run googleadk/multiagent/starter/main.go
```

The starter asks a weather question; the coordinator transfers to the `weather`
specialist, which calls `get_weather` and answers. In the Temporal UI you will see
one `googleadk.InvokeModel` Activity per agent turn plus the `get_weather` Activity.

### Test without a live LLM

`workflow_test.go` scripts each agent's `FakeModel` (coordinator transfers to
`weather`; the weather specialist calls `get_weather` then answers) and asserts the
transfer took effect, so it needs no API key or network:
```bash
go test ./googleadk/multiagent/...
```
