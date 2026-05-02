// Unit tests for BuildTriageRegistry.
//
// Drives the registry directly via Dispatch — bypasses RunWithSession and the
// LLM provider. Asserts that the agent's tool-call sequence produces the
// expected final result.
//
// Mirrors workers/typescript/triage_activity.test.ts and
// workers/python/tests/test_triage_activity.py.
package activities

import (
	"context"
	"strings"
	"testing"
	"time"

	triage "github.com/temporalio/samples-go/toolregistry-incident-triage"
	tr "go.temporal.io/sdk/contrib/toolregistry"
)

func makeAlert() triage.AlertPayload {
	return triage.AlertPayload{
		Status: "firing",
		Labels: map[string]string{
			"alertname": "HighLatencyP99",
			"service":   "api",
			"runbook":   "rollback-or-scale",
		},
		Annotations: map[string]string{
			"summary":     "P99 > 1s",
			"description": "P99 above threshold for 1m.",
		},
		StartsAt: time.Now().UTC(),
	}
}

// makeDeps returns TriageDeps with default mocks; pass overrides as a
// callback that mutates the struct.
func makeDeps(over func(*TriageDeps)) TriageDeps {
	d := TriageDeps{
		MCPListTools: func(baseURL string) ([]MCPToolInfo, error) {
			if strings.Contains(baseURL, "7071") {
				return []MCPToolInfo{{
					Name:        "prometheus_query",
					Description: "instant PromQL query",
					InputSchema: map[string]any{
						"type":       "object",
						"properties": map[string]any{"query": map[string]any{"type": "string"}},
						"required":   []string{"query"},
					},
				}}, nil
			}
			return []MCPToolInfo{{
				Name:        "kubectl_describe",
				Description: "describe a k8s resource",
				InputSchema: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"resource":  map[string]any{"type": "string"},
						"name":      map[string]any{"type": "string"},
						"namespace": map[string]any{"type": "string"},
					},
					"required": []string{"resource", "name"},
				},
			}}, nil
		},
		MCPCallTool: func(_ string, name string, _ map[string]any) (string, error) {
			return "(mocked " + name + ")", nil
		},
		RequestHumanApproval: func(_ context.Context, _ triage.AlertPayload, _ triage.ApprovalRequest) (triage.ApprovalResponse, error) {
			return triage.ApprovalResponse{Decision: "approved", Reason: "default-mock"}, nil
		},
		ExecShellCommand: func(cmd string) (string, string, error) {
			return "(mocked exec: " + cmd + ")", "", nil
		},
	}
	if over != nil {
		over(&d)
	}
	return d
}

// scriptedCall represents one tool the test wants the registry to dispatch.
type scriptedCall struct {
	name  string
	input map[string]any
}

// drive builds the registry and dispatches the scripted tool calls in order,
// then returns (final, sessionResults).
func drive(t *testing.T, deps TriageDeps, calls []scriptedCall) (*triage.TriageResult, []map[string]any) {
	t.Helper()
	session := &tr.AgenticSession{}
	registry, getResult := BuildTriageRegistry(context.Background(), makeAlert(), session, deps)
	for _, c := range calls {
		if _, err := registry.Dispatch(c.name, c.input); err != nil {
			t.Fatalf("dispatch %s: %v", c.name, err)
		}
	}
	return getResult(), session.Results
}

func TestHappyPathResolved(t *testing.T) {
	approvalCalls := 0
	deps := makeDeps(func(d *TriageDeps) {
		d.RequestHumanApproval = func(_ context.Context, _ triage.AlertPayload, _ triage.ApprovalRequest) (triage.ApprovalResponse, error) {
			approvalCalls++
			return triage.ApprovalResponse{Decision: "approved", Reason: "go ahead"}, nil
		}
	})
	action := "kubectl rollout restart deploy/api -n demo-app"

	result, sessionResults := drive(t, deps, []scriptedCall{
		{"prometheus_query", map[string]any{"query": "up{service='api'}"}},
		{"kubectl_describe", map[string]any{"resource": "pod", "name": "api-xyz", "namespace": "demo-app"}},
		{"propose_remediation", map[string]any{"action": action, "justification": "leak; restart reclaims memory"}},
		{"request_human_approval", map[string]any{
			"message": "Restart api?", "diagnosis": "memory leak", "proposedAction": action,
		}},
		{"execute_remediation", map[string]any{"action": action}},
		{"report_resolved", map[string]any{"summary": "restarted; latency normal"}},
	})

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Status != "resolved" {
		t.Errorf("status = %q, want resolved", result.Status)
	}
	if !strings.Contains(result.Summary, "restart") {
		t.Errorf("summary = %q, want contains 'restart'", result.Summary)
	}
	if len(result.Remediations) != 1 || result.Remediations[0].Action != action {
		t.Errorf("remediations = %+v", result.Remediations)
	}
	if approvalCalls != 1 {
		t.Errorf("approval calls = %d, want 1", approvalCalls)
	}

	wantKinds := []string{"remediation", "approval", "executed", "final"}
	if len(sessionResults) != 4 {
		t.Fatalf("session results len = %d, want 4", len(sessionResults))
	}
	for i, want := range wantKinds {
		got := sessionResults[i]["kind"]
		if got != want {
			t.Errorf("session result %d kind = %v, want %s", i, got, want)
		}
	}
}

func TestRejectedApprovalUnresolved(t *testing.T) {
	deps := makeDeps(func(d *TriageDeps) {
		d.RequestHumanApproval = func(_ context.Context, _ triage.AlertPayload, _ triage.ApprovalRequest) (triage.ApprovalResponse, error) {
			return triage.ApprovalResponse{Decision: "rejected", Reason: "off-hours; defer until tomorrow"}, nil
		}
	})

	result, sessionResults := drive(t, deps, []scriptedCall{
		{"propose_remediation", map[string]any{"action": "kubectl scale ...", "justification": "transient"}},
		{"request_human_approval", map[string]any{
			"message": "Scale?", "diagnosis": "transient", "proposedAction": "kubectl scale ...",
		}},
		{"report_unresolved", map[string]any{"summary": "operator deferred"}},
	})

	if result.Status != "unresolved" {
		t.Errorf("status = %q, want unresolved", result.Status)
	}
	if !strings.Contains(result.Summary, "deferred") {
		t.Errorf("summary = %q, want contains 'deferred'", result.Summary)
	}
	// Approval result should be in session results so operator can audit.
	var approval map[string]any
	for _, r := range sessionResults {
		if r["kind"] == "approval" {
			approval = r
			break
		}
	}
	if approval == nil {
		t.Fatal("no approval session result")
	}
	if approval["decision"] != "rejected" {
		t.Errorf("decision = %v, want rejected", approval["decision"])
	}
	if !strings.Contains(approval["reason"].(string), "off-hours") {
		t.Errorf("reason = %v, want contains 'off-hours'", approval["reason"])
	}
}

func TestExecuteRefusesWithoutApproval(t *testing.T) {
	executed := false
	deps := makeDeps(func(d *TriageDeps) {
		d.ExecShellCommand = func(cmd string) (string, string, error) {
			executed = true
			return "", "", nil
		}
	})

	result, _ := drive(t, deps, []scriptedCall{
		// Skip approval — agent tries to execute directly.
		{"execute_remediation", map[string]any{"action": "rm -rf /"}},
		{"report_unresolved", map[string]any{"summary": "tried to skip approval"}},
	})

	if result.Status != "unresolved" {
		t.Errorf("status = %q, want unresolved", result.Status)
	}
	if executed {
		t.Error("ExecShellCommand should not have been called")
	}
}

func TestExecuteRefusesWhenActionDoesNotMatch(t *testing.T) {
	var executedCmd string
	deps := makeDeps(func(d *TriageDeps) {
		d.RequestHumanApproval = func(_ context.Context, _ triage.AlertPayload, _ triage.ApprovalRequest) (triage.ApprovalResponse, error) {
			return triage.ApprovalResponse{Decision: "approved", Reason: "ok"}, nil
		}
		d.ExecShellCommand = func(cmd string) (string, string, error) {
			executedCmd = cmd
			return "ran", "", nil
		}
	})

	result, _ := drive(t, deps, []scriptedCall{
		{"propose_remediation", map[string]any{"action": "kubectl restart api", "justification": "x"}},
		{"request_human_approval", map[string]any{
			"message": "Restart?", "diagnosis": "x", "proposedAction": "kubectl restart api",
		}},
		// Agent attempts a DIFFERENT action than what was approved.
		{"execute_remediation", map[string]any{"action": "kubectl scale deploy/api --replicas=10"}},
		{"report_unresolved", map[string]any{"summary": "guard tripped"}},
	})

	if result.Status != "unresolved" {
		t.Errorf("status = %q, want unresolved", result.Status)
	}
	if executedCmd != "" {
		t.Errorf("ExecShellCommand was called with %q; should not have been called", executedCmd)
	}
}

func TestMCPToolsRegistered(t *testing.T) {
	deps := makeDeps(nil)
	session := &tr.AgenticSession{}
	registry, _ := BuildTriageRegistry(context.Background(), makeAlert(), session, deps)
	defs := registry.Defs()
	names := make(map[string]bool)
	for _, d := range defs {
		names[d.Name] = true
	}
	for _, want := range []string{
		"prometheus_query", "kubectl_describe",
		"propose_remediation", "request_human_approval",
		"execute_remediation", "report_resolved", "report_unresolved",
	} {
		if !names[want] {
			t.Errorf("missing tool: %s", want)
		}
	}
}

func TestMCPDispatchForwardsToSidecar(t *testing.T) {
	type call struct {
		url, name string
		args      map[string]any
	}
	var calls []call
	deps := makeDeps(func(d *TriageDeps) {
		d.MCPCallTool = func(url, name string, args map[string]any) (string, error) {
			calls = append(calls, call{url, name, args})
			return "result for " + name, nil
		}
	})

	_, _ = drive(t, deps, []scriptedCall{
		{"prometheus_query", map[string]any{"query": "up{}"}},
		{"report_unresolved", map[string]any{"summary": "test"}},
	})

	if len(calls) != 1 {
		t.Fatalf("got %d MCP calls, want 1", len(calls))
	}
	if calls[0].name != "prometheus_query" {
		t.Errorf("name = %s, want prometheus_query", calls[0].name)
	}
	if !strings.Contains(calls[0].url, "7071") {
		t.Errorf("url = %s, want contains 7071", calls[0].url)
	}
	if calls[0].args["query"] != "up{}" {
		t.Errorf("args = %v", calls[0].args)
	}
}
