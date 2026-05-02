# Go: incident-triage tool-registry sample

Demonstrates `go.temporal.io/sdk/contrib/toolregistry` end-to-end: long-running `RunWithSession` activity, MCP HTTP integration, human-in-the-loop via companion workflow, and a testable activity refactor.

## What's here

| File | Purpose |
|---|---|
| `types.go` | `AlertPayload`, `TriageResult`, `ApprovalRequest`, `ApprovalResponse` types. |
| `activities/triage.go` | The activity. `TriageDeps` struct, `BuildTriageRegistry(alert, session, deps)` returning `(*toolregistry.ToolRegistry, func() *TriageResult)`, and the activity entrypoint. |
| `activities/triage_test.go` | Unit tests demonstrating `MockProvider` + `TriageDeps` pattern. |
| `workflows/triage.go`, `workflows/approval.go` | Workflows. |
| `cmd/worker/main.go`, `cmd/client/main.go` | Worker entrypoint and demo client. |

## Run

```bash
temporal server start-dev          # separate terminal

export ANTHROPIC_API_KEY=sk-ant-...
export PROM_MCP=http://localhost:7070/mcp
export K8S_MCP=http://localhost:7071/mcp

go run ./cmd/worker                # worker
go run ./cmd/client                # client
go test ./...                      # tests
```
