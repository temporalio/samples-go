# Google ADK model usage metrics

This sample records successful Google ADK model calls, model errors, and token usage from `LLMResponse.UsageMetadata`. The ADK callbacks emit through Temporal's Workflow metrics handler, which suppresses emissions during replay, and tag each counter with fixed agent and model names.

| Counter (as scraped) | Source |
| --- | --- |
| `google_adk_model_calls_total` | One per successful model turn |
| `google_adk_model_errors_total` | One per failed model turn |
| `google_adk_input_tokens_total` | `PromptTokenCount` |
| `google_adk_output_tokens_total` | `CandidatesTokenCount` |
| `google_adk_reasoning_tokens_total` | `ThoughtsTokenCount` |
| `google_adk_cached_input_tokens_total` | `CachedContentTokenCount` |
| `google_adk_total_tokens_total` | `TotalTokenCount` |

Prometheus adds the `_total` suffix and sanitizes tag values. The SDK also adds a `workflow_type` tag, so a scraped line looks like `google_adk_model_calls_total{agent="metrics_assistant",model="gemini_2_5_flash",workflow_type="AgentWorkflow"} 1`.

Replay suppression makes these metrics at-least-once rather than exactly-once. A Workflow Task retry re-executes its live segment and can record again, so use these as usage signals rather than billing records.

Start a local Temporal Service:

```sh
temporal server start-dev
```

Set a Gemini API key and run the worker:

```sh
export GEMINI_API_KEY=...
go run googleadk/metrics/worker/main.go
```

In another terminal, start the Workflow:

```sh
go run googleadk/metrics/starter/main.go
```

Inspect the Prometheus endpoint:

```sh
curl -s http://127.0.0.1:9090/metrics | grep google_adk_
```
