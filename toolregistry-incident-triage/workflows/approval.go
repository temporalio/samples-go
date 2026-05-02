// Package workflows hosts the Temporal workflow definitions for the triage flow.
package workflows

import (
	triage "github.com/temporalio/samples-go/toolregistry-incident-triage"
	"go.temporal.io/sdk/workflow"
)

const (
	ApprovalRequestSignal  = "approval-request"
	ApprovalDecisionSignal = "approval-decision"
	PendingApprovalQuery   = "pending-approval"
)

// ApprovalWorkflow is the companion HITL workflow.
//
// The triage agent's request_human_approval tool calls SignalWithStartWorkflow
// against a deterministic ID per alert group. This workflow stores the latest
// agent request, exposes it as a query, and returns the operator's decision.
//
// Same shape as the TypeScript and Python references.
func ApprovalWorkflow(ctx workflow.Context, _key string) (*triage.ApprovalResponse, error) {
	logger := workflow.GetLogger(ctx)

	var request *triage.ApprovalRequest
	var response *triage.ApprovalResponse

	if err := workflow.SetQueryHandler(ctx, PendingApprovalQuery, func() (*triage.ApprovalRequest, error) {
		return request, nil
	}); err != nil {
		return nil, err
	}

	requestCh := workflow.GetSignalChannel(ctx, ApprovalRequestSignal)
	decisionCh := workflow.GetSignalChannel(ctx, ApprovalDecisionSignal)

	// Wait for the agent's request, then the operator's decision.
	// LLM retry: re-attached requests overwrite prior state — operator only
	// sees the latest version, since the agent may refine its ask.
	for response == nil {
		selector := workflow.NewSelector(ctx)
		selector.AddReceive(requestCh, func(c workflow.ReceiveChannel, _ bool) {
			var req triage.ApprovalRequest
			c.Receive(ctx, &req)
			request = &req
			logger.Info("approval request received", "message", req.Message)
		})
		selector.AddReceive(decisionCh, func(c workflow.ReceiveChannel, _ bool) {
			var res triage.ApprovalResponse
			c.Receive(ctx, &res)
			response = &res
		})
		selector.Select(ctx)
	}

	return response, nil
}
