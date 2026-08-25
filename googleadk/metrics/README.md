# Google ADK model usage metrics

This sample records successful Google ADK model calls and their input, output, reasoning, cached-input, and total token counts from `LLMResponse.UsageMetadata`. The ADK callback uses Temporal's Workflow metrics handler, which suppresses emissions during replay, and tags each counter with fixed agent and model names.

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
curl http://127.0.0.1:9090/metrics | rg 'google_adk_'
```
