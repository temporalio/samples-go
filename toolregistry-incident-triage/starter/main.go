// Client CLI for the Go triage workers.
//
// Usage:
//
//	client approve <workflow-id> <reason>
//	client reject  <workflow-id> <reason>
//	client trigger <alertname> <service>    # post a synthetic alert (skips webhook)
//
// To list pending approval workflows, use the Temporal CLI directly:
//
//	temporal workflow list --query 'WorkflowType="ApprovalWorkflow" AND ExecutionStatus="Running"'
//
// (The Go SDK's list-workflow API surface varies across versions; the Temporal CLI
// is the most reliable cross-version way to enumerate.)
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	triage "github.com/temporalio/samples-go/toolregistry-incident-triage"
	"github.com/temporalio/samples-go/toolregistry-incident-triage/workflows"
	"go.temporal.io/sdk/client"
)

func makeClient() client.Client {
	address := mustEnv("TEMPORAL_ADDRESS")
	namespace := mustEnv("TEMPORAL_NAMESPACE")
	apiKey := mustEnv("TEMPORAL_API_KEY")
	c, err := client.Dial(client.Options{
		HostPort:    address,
		Namespace:   namespace,
		Credentials: client.NewAPIKeyStaticCredentials(apiKey),
	})
	if err != nil {
		log.Fatalf("temporal dial: %v", err)
	}
	return c
}

func mustEnv(name string) string {
	v := os.Getenv(name)
	if v == "" {
		log.Fatalf("missing env var: %s", name)
	}
	return v
}

func decide(decision, workflowID, reason string) {
	c := makeClient()
	defer c.Close()
	ctx := context.Background()
	if err := c.SignalWorkflow(ctx, workflowID, "", workflows.ApprovalDecisionSignal,
		triage.ApprovalResponse{Decision: decision, Reason: reason}); err != nil {
		log.Fatalf("signal: %v", err)
	}
	fmt.Printf("signaled %s: %s — %s\n", workflowID, decision, reason)
}

func trigger(alertname, service string) {
	c := makeClient()
	defer c.Close()
	taskQueue := os.Getenv("TEMPORAL_TASK_QUEUE")
	if taskQueue == "" {
		taskQueue = "triage-go"
	}
	wfID := fmt.Sprintf("triage-%s-%s", strings.ToLower(alertname), strings.ToLower(service))
	alert := triage.AlertPayload{
		Status: "firing",
		Labels: map[string]string{"alertname": alertname, "service": service, "severity": "critical", "runbook": "synthetic"},
		Annotations: map[string]string{
			"summary":     fmt.Sprintf("Synthetic test alert for %s", service),
			"description": "Triggered manually via client CLI to exercise the triage flow.",
		},
		StartsAt: time.Now().UTC(),
	}
	ctx := context.Background()
	run, err := c.SignalWithStartWorkflow(ctx, wfID, workflows.AlertUpdateSignal, alert,
		client.StartWorkflowOptions{ID: wfID, TaskQueue: taskQueue},
		"IncidentTriageWorkflow", alert)
	if err != nil {
		log.Fatalf("signal-with-start: %v", err)
	}
	fmt.Printf("started triage workflow: %s (run %s) on %s\n", wfID, run.GetRunID(), taskQueue)
}

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		log.Fatalln("Usage: client <approve|reject|trigger> ...")
	}
	cmd, rest := args[0], args[1:]
	switch cmd {
	case "approve":
		if len(rest) < 2 {
			log.Fatalln("Usage: client approve <wfid> <reason>")
		}
		decide("approved", rest[0], strings.Join(rest[1:], " "))
	case "reject":
		if len(rest) < 2 {
			log.Fatalln("Usage: client reject <wfid> <reason>")
		}
		decide("rejected", rest[0], strings.Join(rest[1:], " "))
	case "trigger":
		if len(rest) < 2 {
			log.Fatalln("Usage: client trigger <alertname> <service>")
		}
		trigger(rest[0], rest[1])
	default:
		log.Fatalf("Unknown command: %s", cmd)
	}
}
