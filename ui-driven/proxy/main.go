package proxy

import (
	"fmt"

	"go.temporal.io/sdk/workflow"
)

const proxyRequestSignalName = "proxy-request-signal"
const proxyResponseSignalName = "proxy-response-signal"

type proxySignalRequest struct {
	Key               string
	Value             string
	CallingWorkflowId string
}

type proxySignalResponse struct {
	Key   string
	Value string
	Error string
}

func SendErrorResponse(ctx workflow.Context, id string, err error) error {
	logger := workflow.GetLogger(ctx)

	logger.Info("Sending error response", id)

	return workflow.SignalExternalWorkflow(
		ctx,
		id,
		"",
		proxyResponseSignalName,
		proxySignalResponse{Error: err.Error()},
	).Get(ctx, nil)
}

func SendResponse(ctx workflow.Context, id string, key string, value string) error {
	logger := workflow.GetLogger(ctx)

	logger.Info("Sending response", id)

	return workflow.SignalExternalWorkflow(
		ctx,
		id,
		"",
		proxyResponseSignalName,
		proxySignalResponse{Key: key, Value: value},
	).Get(ctx, nil)
}

func SendRequest(ctx workflow.Context, targetWorkflowID string, key string, value string) error {
	logger := workflow.GetLogger(ctx)

	workflowID := workflow.GetInfo(ctx).WorkflowExecution.ID

	logger.Info("Sending request", targetWorkflowID, workflowID)

	return workflow.SignalExternalWorkflow(
		ctx,
		targetWorkflowID,
		"",
		proxyRequestSignalName,
		proxySignalRequest{
			CallingWorkflowId: workflowID,
			Key:               key,
			Value:             value,
		},
	).Get(ctx, nil)
}

func ReceiveResponse(ctx workflow.Context) (string, string, error) {
	logger := workflow.GetLogger(ctx)

	var res proxySignalResponse

	ch := workflow.GetSignalChannel(ctx, proxyResponseSignalName)

	logger.Info("Waiting for response")

	ch.Receive(ctx, &res)

	logger.Info("Received response")

	if res.Error != "" {
		return "", "", fmt.Errorf("%s", res.Error)
	}

	return res.Key, res.Value, nil
}

func ReceiveRequest(ctx workflow.Context) (string, string, string) {
	logger := workflow.GetLogger(ctx)

	var req proxySignalRequest

	ch := workflow.GetSignalChannel(ctx, proxyRequestSignalName)

	logger.Info("Waiting for request")

	ch.Receive(ctx, &req)

	logger.Info("Received request")

	return req.CallingWorkflowId, req.Key, req.Value
}
