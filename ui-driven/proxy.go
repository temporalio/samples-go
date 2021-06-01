package uidriven

import "go.temporal.io/sdk/workflow"

const UIRequestSignalName = "ui-request-signal"
const UIResponseSignalName = "ui-response-signal"

type UISignalRequest struct {
	Stage             string
	Value             string
	CallingWorkflowId string
}

type UISignalResponse struct {
	Error error
	Stage string
}

func SendErrorResponseToUI(ctx workflow.Context, req UISignalRequest, err error) error {
	return workflow.SignalExternalWorkflow(
		ctx,
		req.CallingWorkflowId,
		"",
		UIRequestSignalName,
		UISignalResponse{Error: err},
	).Get(ctx, nil)
}

func SendResponseToUI(ctx workflow.Context, req UISignalRequest, stage string) error {
	return workflow.SignalExternalWorkflow(
		ctx,
		req.CallingWorkflowId,
		"",
		UIRequestSignalName,
		UISignalResponse{Stage: stage},
	).Get(ctx, nil)
}

func SendRequestToOrderWorkflow(ctx workflow.Context, orderWorkflowID string, stage string, value string) error {
	workflowID := workflow.GetInfo(ctx).WorkflowExecution.ID

	return workflow.SignalExternalWorkflow(
		ctx,
		orderWorkflowID,
		"",
		UIRequestSignalName,
		UISignalRequest{
			CallingWorkflowId: workflowID,
			Stage:             stage,ß
			Value:             value,
		},
	).Get(ctx, nil)
}

func ReceiveResponseFromOrderWorkflow(ctx workflow.Context) (UISignalResponse, error) {
	var res UISignalResponse

	uiCh := workflow.GetSignalChannel(ctx, UIResponseSignalName)

	uiCh.Receive(ctx, &res)

	return res, res.Error
}

func ReceiveRequestFromUI(ctx workflow.Context) UISignalRequest {
	var req UISignalRequest

	uiCh := workflow.GetSignalChannel(ctx, UIRequestSignalName)

	uiCh.Receive(ctx, &req)

	return req
}
