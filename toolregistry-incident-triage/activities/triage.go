// Package activities implements triage_incident_activity, the agentic loop.
//
// Mirrors workers/typescript/activities/triage.ts and workers/python/triage_activity.py:
//   - Pulls Prometheus + Kubernetes tools from MCP sidecars (localhost:7071/7072)
//     via JSON-RPC over HTTP, registers them on the ToolRegistry.
//   - Defines per-language tools: propose_remediation, request_human_approval,
//     execute_remediation, report_resolved, report_unresolved.
//   - Opens an AgenticSession via RunWithSession, runs the loop, returns the result.
//
// Structure for testability:
//   - BuildTriageRegistry returns the (registry, getResult) pair. Pure-ish:
//     takes all I/O dependencies as injected callables so unit tests can
//     substitute them.
//   - TriageIncidentActivity opens the session and composes the call.
package activities

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	triage "github.com/temporalio/samples-go/toolregistry-incident-triage"
	"github.com/temporalio/samples-go/toolregistry-incident-triage/workflows"
	"go.temporal.io/sdk/client"
	tr "go.temporal.io/sdk/contrib/toolregistry"
)

const SystemPrompt = `You are an SRE on-call agent triaging a production alert.

You have these tools (sourced from MCP sidecars + per-language helpers):
  - prometheus_query(query)            instant PromQL query
  - prometheus_query_range(query, start, end, step)
  - prometheus_alerts()                what is currently firing
  - kubectl_get(resource, namespace?)  list K8s resources
  - kubectl_describe(resource, name, namespace?)
  - kubectl_logs(pod, namespace, tail?)
  - propose_remediation(action, justification)   record but do NOT execute
  - request_human_approval(message, diagnosis, proposedAction)
                                       blocks until operator says approve|reject
  - execute_remediation(action)        ONLY callable AFTER approval was approved.
                                       Pass the same action you got approved.
  - report_resolved(summary)           ends the loop with status=resolved
  - report_unresolved(summary)         ends the loop with status=unresolved

Workflow:
  1. Read the alert. Use prometheus_query to confirm the symptom is currently true.
  2. Use kubectl_get/describe/logs and prometheus_query_range to find root cause.
  3. propose_remediation with a specific action.
  4. request_human_approval, attaching your diagnosis and the proposed action.
  5. If approved: execute_remediation, then prometheus_query to verify, then report_resolved.
  6. If rejected: report_unresolved with the operator's reason.

Be terse. Conversation history is heartbeated to Temporal — keep tool inputs short.`

// TriageDeps holds injectable I/O for the activity. Tests substitute their own.
type TriageDeps struct {
	MCPListTools         func(baseURL string) ([]MCPToolInfo, error)
	MCPCallTool          func(baseURL, name string, args map[string]any) (string, error)
	RequestHumanApproval func(alert triage.AlertPayload, req triage.ApprovalRequest) (triage.ApprovalResponse, error)
	ExecShellCommand     func(cmd string) (string, string, error)
}

// MCPToolInfo is the subset of an MCP tool catalog entry we care about.
type MCPToolInfo struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
}

func defaultMCPListTools(baseURL string) ([]MCPToolInfo, error) {
	body, err := mcpRPC(baseURL, "tools/list", nil)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Result struct{ Tools []MCPToolInfo } `json:"result"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}
	return resp.Result.Tools, nil
}

func defaultMCPCallTool(baseURL, name string, args map[string]any) (string, error) {
	body, err := mcpRPC(baseURL, "tools/call", map[string]any{"name": name, "arguments": args})
	if err != nil {
		return fmt.Sprintf("MCP error: %v", err), nil
	}
	var resp struct {
		Result struct {
			Content []struct{ Text string } `json:"content"`
		} `json:"result"`
		Error *struct{ Message string }
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", err
	}
	if resp.Error != nil {
		return fmt.Sprintf("MCP error: %s", resp.Error.Message), nil
	}
	parts := make([]string, 0, len(resp.Result.Content))
	for _, c := range resp.Result.Content {
		parts = append(parts, c.Text)
	}
	return strings.Join(parts, "\n"), nil
}

func mcpRPC(baseURL, method string, params any) ([]byte, error) {
	payload, _ := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"id":      time.Now().UnixNano(),
		"method":  method,
		"params":  params,
	})
	req, _ := http.NewRequest("POST", baseURL, bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	c := &http.Client{Timeout: 30 * time.Second}
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func defaultExecShellCommand(cmd string) (string, string, error) {
	c := exec.Command("sh", "-c", cmd)
	var stdout, stderr bytes.Buffer
	c.Stdout = &stdout
	c.Stderr = &stderr
	err := c.Run()
	return stdout.String(), stderr.String(), err
}

// DefaultDeps returns deps wired to the real MCP HTTP client, real shell exec,
// and real Temporal client for HITL.
func DefaultDeps() TriageDeps {
	return TriageDeps{
		MCPListTools:         defaultMCPListTools,
		MCPCallTool:          defaultMCPCallTool,
		RequestHumanApproval: realRequestHumanApproval,
		ExecShellCommand:     defaultExecShellCommand,
	}
}

func envOrDefault(name, def string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}
	return def
}

// BuildTriageRegistry builds a populated *ToolRegistry plus a getResult()
// accessor for the final verdict. Pure modulo deps.
func BuildTriageRegistry(
	alert triage.AlertPayload,
	session *tr.AgenticSession,
	deps TriageDeps,
) (*tr.ToolRegistry, func() *triage.TriageResult) {
	registry := tr.NewToolRegistry()
	promMCP := envOrDefault("MCP_PROMETHEUS_URL", "http://localhost:7071/")
	k8sMCP := envOrDefault("MCP_KUBERNETES_URL", "http://localhost:7072/")

	// MCP-sourced tools.
	for _, baseURL := range []string{promMCP, k8sMCP} {
		tools, err := deps.MCPListTools(baseURL)
		if err != nil {
			continue
		}
		// Capture loop variables.
		url := baseURL
		for _, t := range tools {
			name := t.Name
			schema := t.InputSchema
			if schema == nil {
				schema = map[string]any{"type": "object"}
			}
			registry.Register(
				tr.ToolDef{Name: name, Description: t.Description, InputSchema: schema},
				func(input map[string]any) (string, error) {
					return deps.MCPCallTool(url, name, input)
				},
			)
		}
	}

	// Per-language tools.
	var remediations []triage.ProposedRemediation
	var approvedAction string
	var final *triage.TriageResult

	registry.Register(
		tr.ToolDef{
			Name:        "propose_remediation",
			Description: "Record a remediation you would apply. Does NOT execute it.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"action":        map[string]any{"type": "string"},
					"justification": map[string]any{"type": "string"},
				},
				"required": []string{"action", "justification"},
			},
		},
		func(inp map[string]any) (string, error) {
			r := triage.ProposedRemediation{
				Action:        fmt.Sprint(inp["action"]),
				Justification: fmt.Sprint(inp["justification"]),
			}
			remediations = append(remediations, r)
			session.Results = append(session.Results, map[string]any{
				"kind": "remediation", "action": r.Action, "justification": r.Justification,
			})
			return "recorded", nil
		},
	)

	registry.Register(
		tr.ToolDef{
			Name:        "request_human_approval",
			Description: "Block until operator decides. Returns JSON {decision, reason}.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"message":        map[string]any{"type": "string"},
					"diagnosis":      map[string]any{"type": "string"},
					"proposedAction": map[string]any{"type": "string"},
				},
				"required": []string{"message", "diagnosis", "proposedAction"},
			},
		},
		func(inp map[string]any) (string, error) {
			req := triage.ApprovalRequest{
				Message:        fmt.Sprint(inp["message"]),
				Diagnosis:      fmt.Sprint(inp["diagnosis"]),
				ProposedAction: fmt.Sprint(inp["proposedAction"]),
			}
			resp, err := deps.RequestHumanApproval(alert, req)
			if err != nil {
				return "", err
			}
			if resp.Decision == "approved" {
				approvedAction = req.ProposedAction
			}
			session.Results = append(session.Results, map[string]any{
				"kind": "approval", "decision": resp.Decision, "reason": resp.Reason,
			})
			b, _ := json.Marshal(resp)
			return string(b), nil
		},
	)

	registry.Register(
		tr.ToolDef{
			Name:        "execute_remediation",
			Description: "Execute the previously-approved action. Errors if no approval has been granted.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"action": map[string]any{"type": "string"},
				},
				"required": []string{"action"},
			},
		},
		func(inp map[string]any) (string, error) {
			action := fmt.Sprint(inp["action"])
			if approvedAction == "" {
				return "ERROR: no approval has been granted. Call request_human_approval first.", nil
			}
			if action != approvedAction {
				return fmt.Sprintf("ERROR: requested action does not match approved action. Approved: %s", approvedAction), nil
			}
			stdout, stderr, err := deps.ExecShellCommand(action)
			if err != nil {
				return fmt.Sprintf("EXEC ERROR: %v", err), nil
			}
			session.Results = append(session.Results, map[string]any{
				"kind": "executed", "action": action,
				"stdout": clip(stdout, 2000), "stderr": clip(stderr, 2000),
			})
			out := stdout
			if out == "" {
				out = stderr
			}
			if out == "" {
				out = "ok"
			}
			return clip(out, 4000), nil
		},
	)

	registry.Register(
		tr.ToolDef{
			Name:        "report_resolved",
			Description: "Ends the loop with status=resolved.",
			InputSchema: map[string]any{
				"type":       "object",
				"properties": map[string]any{"summary": map[string]any{"type": "string"}},
				"required":   []string{"summary"},
			},
		},
		func(inp map[string]any) (string, error) {
			r := &triage.TriageResult{
				Status:       "resolved",
				Summary:      fmt.Sprint(inp["summary"]),
				Remediations: append([]triage.ProposedRemediation(nil), remediations...),
			}
			final = r
			session.Results = append(session.Results, map[string]any{"kind": "final", "result": r})
			return "ok", nil
		},
	)

	registry.Register(
		tr.ToolDef{
			Name:        "report_unresolved",
			Description: "Ends the loop with status=unresolved.",
			InputSchema: map[string]any{
				"type":       "object",
				"properties": map[string]any{"summary": map[string]any{"type": "string"}},
				"required":   []string{"summary"},
			},
		},
		func(inp map[string]any) (string, error) {
			r := &triage.TriageResult{
				Status:       "unresolved",
				Summary:      fmt.Sprint(inp["summary"]),
				Remediations: append([]triage.ProposedRemediation(nil), remediations...),
			}
			final = r
			session.Results = append(session.Results, map[string]any{"kind": "final", "result": r})
			return "ok", nil
		},
	)

	return registry, func() *triage.TriageResult { return final }
}

func BuildPrompt(alert triage.AlertPayload) string {
	get := func(m map[string]string, k, d string) string {
		if v, ok := m[k]; ok && v != "" {
			return v
		}
		return d
	}
	return fmt.Sprintf(
		"Alert fired: %s on %s.\nSummary: %s\nDescription: %s\nRunbook hint: %s\n\nInvestigate, propose, get approval, and either fix or report unresolved.",
		get(alert.Labels, "alertname", "unknown"),
		get(alert.Labels, "service", "unknown"),
		get(alert.Annotations, "summary", "(none)"),
		get(alert.Annotations, "description", "(none)"),
		get(alert.Labels, "runbook", "(none)"),
	)
}

func clip(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

// TriageIncidentActivity is the entrypoint Temporal calls for the triage flow.
func TriageIncidentActivity(ctx context.Context, alert triage.AlertPayload) (*triage.TriageResult, error) {
	deps := DefaultDeps()
	var result *triage.TriageResult
	err := tr.RunWithSession(ctx, func(ctx context.Context, session *tr.AgenticSession) error {
		registry, getResult := BuildTriageRegistry(alert, session, deps)
		cfg := tr.AnthropicConfig{APIKey: os.Getenv("ANTHROPIC_API_KEY")}
		provider := tr.NewAnthropicProvider(cfg, registry, SystemPrompt)
		if err := session.RunToolLoop(ctx, provider, registry, BuildPrompt(alert)); err != nil {
			return err
		}
		result = getResult()
		if result == nil {
			return fmt.Errorf("agent ended the loop without calling report_resolved or report_unresolved")
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// realRequestHumanApproval signal-with-starts ApprovalWorkflow against a
// deterministic ID per alert group and waits for the result.
func realRequestHumanApproval(alert triage.AlertPayload, req triage.ApprovalRequest) (triage.ApprovalResponse, error) {
	apiKey := os.Getenv("TEMPORAL_API_KEY")
	address := os.Getenv("TEMPORAL_ADDRESS")
	namespace := os.Getenv("TEMPORAL_NAMESPACE")
	taskQueue := envOrDefault("TEMPORAL_TASK_QUEUE", "triage-go")
	if apiKey == "" || address == "" || namespace == "" {
		return triage.ApprovalResponse{}, fmt.Errorf("missing TEMPORAL_ADDRESS/NAMESPACE/API_KEY")
	}

	c, err := client.Dial(client.Options{
		HostPort:    address,
		Namespace:   namespace,
		Credentials: client.NewAPIKeyStaticCredentials(apiKey),
	})
	if err != nil {
		return triage.ApprovalResponse{}, err
	}
	defer c.Close()

	key := fmt.Sprintf("%s-%s",
		strings.ToLower(envOrDefault(alert.Labels["alertname"], "unknown")),
		strings.ToLower(envOrDefault(alert.Labels["service"], "unknown")),
	)
	approvalWfID := "approval-" + key

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Hour)
	defer cancel()

	run, err := c.SignalWithStartWorkflow(
		ctx,
		approvalWfID,
		workflows.ApprovalRequestSignal,
		req,
		client.StartWorkflowOptions{ID: approvalWfID, TaskQueue: taskQueue},
		"ApprovalWorkflow",
		key,
	)
	if err != nil {
		return triage.ApprovalResponse{}, err
	}

	var resp triage.ApprovalResponse
	if err := run.Get(ctx, &resp); err != nil {
		return triage.ApprovalResponse{}, err
	}
	return resp, nil
}
