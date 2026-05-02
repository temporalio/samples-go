package workflows

import (
	"time"

	triage "github.com/temporalio/samples-go/toolregistry-incident-triage"
	"go.temporal.io/sdk/workflow"
)

const (
	AlertUpdateSignal   = "alert-update"
	CurrentAlertQuery   = "current-alert"
	TriageResultQuery   = "triage-result"
	TriageActivityName  = "triage_incident_activity"
)

// IncidentTriageWorkflow delegates the agentic loop to a single activity.
//
// Workflow ID is set deterministically by the webhook receiver
// (triage-${alertname}-${service}), so re-fires from AlertManager re-attach
// to the running workflow rather than spawning a new one.
func IncidentTriageWorkflow(ctx workflow.Context, initialAlert triage.AlertPayload) (*triage.TriageResult, error) {
	currentAlert := initialAlert
	var result *triage.TriageResult

	if err := workflow.SetQueryHandler(ctx, CurrentAlertQuery, func() (triage.AlertPayload, error) {
		return currentAlert, nil
	}); err != nil {
		return nil, err
	}
	if err := workflow.SetQueryHandler(ctx, TriageResultQuery, func() (*triage.TriageResult, error) {
		return result, nil
	}); err != nil {
		return nil, err
	}

	// Listen for alert-update signals (webhook re-fires).
	updateCh := workflow.GetSignalChannel(ctx, AlertUpdateSignal)
	workflow.Go(ctx, func(ctx workflow.Context) {
		for {
			var alert triage.AlertPayload
			more := updateCh.Receive(ctx, &alert)
			if !more {
				return
			}
			currentAlert = alert
		}
	})

	// agenticHitl-shaped timeouts (matches lexicon-temporal's profile).
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 8 * time.Hour,
		HeartbeatTimeout:    120 * time.Second,
		// AgenticSession heartbeat is the resume mechanism.
		RetryPolicy: nil, // 1 attempt
	})

	if err := workflow.ExecuteActivity(ctx, TriageActivityName, currentAlert).Get(ctx, &result); err != nil {
		return nil, err
	}
	return result, nil
}
