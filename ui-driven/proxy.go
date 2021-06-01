package uidriven

import (
	"fmt"

	"go.temporal.io/sdk/workflow"
)

const UIRequestSignalName = "ui-request-signal"
const UIResponseSignalName = "ui-response-signal"

type UISignalRequest struct {
	Stage             string
	Value             string
	CallingWorkflowId string
}

type UISignalResponse struct {
	Error string
	Stage string
}

func SendErrorResponseToUI(ctx workflow.Context, req UISignalRequest, err error) error {
	logger := workflow.GetLogger(ctx)

	logger.Info("Sending error to UI workflow", req.CallingWorkflowId)

	return workflow.SignalExternalWorkflow(
		ctx,
		req.CallingWorkflowId,
		"",
		UIResponseSignalName,
		UISignalResponse{Error: err.Error()},
	).Get(ctx, nil)
}

func SendResponseToUI(ctx workflow.Context, req UISignalRequest, stage string) error {
	logger := workflow.GetLogger(ctx)

	logger.Info("Sending response to UI workflow", req.CallingWorkflowId)

	return workflow.SignalExternalWorkflow(
		ctx,
		req.CallingWorkflowId,
		"",
		UIResponseSignalName,
		UISignalResponse{Stage: stage},
	).Get(ctx, nil)
}

func SendRequestToOrderWorkflow(ctx workflow.Context, orderWorkflowID string, stage string, value string) error {
	logger := workflow.GetLogger(ctx)

	workflowID := workflow.GetInfo(ctx).WorkflowExecution.ID

	logger.Info("Sending request to order workflow", orderWorkflowID, workflowID)

	return workflow.SignalExternalWorkflow(
		ctx,
		orderWorkflowID,
		"",
		UIRequestSignalName,
		UISignalRequest{
			CallingWorkflowId: workflowID,
			Stage:             stage,
			Value:             value,
		},
	).Get(ctx, nil)
}

func ReceiveResponseFromOrderWorkflow(ctx workflow.Context) (UISignalResponse, error) {
	logger := workflow.GetLogger(ctx)

	var res UISignalResponse

	uiCh := workflow.GetSignalChannel(ctx, UIResponseSignalName)

	logger.Info("Waiting for response from order workflow")

	uiCh.Receive(ctx, &res)

	logger.Info("Received response from order workflow")

	if res.Error != "" {
		return res, fmt.Errorf("%s", res.Error)
	}

	return res, nil
}

func ReceiveRequestFromUI(ctx workflow.Context) UISignalRequest {
	logger := workflow.GetLogger(ctx)

	var req UISignalRequest

	uiCh := workflow.GetSignalChannel(ctx, UIRequestSignalName)

	logger.Info("Waiting for response from UI workflow")

	uiCh.Receive(ctx, &req)

	logger.Info("Received response from UI workflow")

	return req
}
