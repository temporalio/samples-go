package uidriven

import (
	"context"
	"fmt"
	"time"

	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

const (
	RegisterStage = "register"
	SizeStage     = "size"
	ColorStage    = "color"
	CompleteStage = "complete"
)

type tshirtOrder struct {
	Email string
	Size  string
	Color string
}

type OrderStatus struct {
	OrderID string
	Stage   string
}

// Workflow is a workflow driven by interaction from a UI.
func OrderWorkflow(ctx workflow.Context, email string) (string, error) {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	order := tshirtOrder{}

	err := order.execute(ctx)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%v", order), nil
}

func (order *tshirtOrder) execute(ctx workflow.Context) error {
	// Loop until we receive a valid email
	for {
		req := ReceiveRequestFromUI(ctx)

		if req.Stage != RegisterStage {
			err := SendErrorResponseToUI(ctx, req, fmt.Errorf("unexpected stage: %v", req.Stage))
			if err != nil {
				return err
			}

			continue
		}

		err := workflow.ExecuteActivity(ctx, RegisterEmail, req.Value).Get(ctx, nil)
		if err != nil {
			err := SendErrorResponseToUI(ctx, req, err)
			if err != nil {
				return err
			}

			continue
		}

		order.Email = req.Value

		err = SendResponseToUI(ctx, req, SizeStage)
		if err != nil {
			return err
		}

		break
	}

	// Loop until we receive a valid size
	for {
		req := ReceiveRequestFromUI(ctx)

		if req.Stage != SizeStage {
			err := SendErrorResponseToUI(ctx, req, fmt.Errorf("unexpected stage: %v", req.Stage))
			if err != nil {
				return err
			}

			continue
		}

		err := workflow.ExecuteActivity(ctx, ValidateSize, req.Value).Get(ctx, nil)
		if err != nil {
			err := SendErrorResponseToUI(ctx, req, err)
			if err != nil {
				return err
			}

			continue
		}

		order.Size = req.Value

		err = SendResponseToUI(ctx, req, ColorStage)
		if err != nil {
			return err
		}

		break
	}

	// Loop until we receive a valid color
	for {
		req := ReceiveRequestFromUI(ctx)

		if req.Stage != ColorStage {
			err := SendErrorResponseToUI(ctx, req, fmt.Errorf("unexpected stage: %v", req.Stage))
			if err != nil {
				return err
			}

			continue
		}

		err := workflow.ExecuteActivity(ctx, ValidateColor, req.Value).Get(ctx, nil)
		if err != nil {
			err := SendErrorResponseToUI(ctx, req, err)
			if err != nil {
				return err
			}

			continue
		}

		order.Color = req.Value

		// Tell the UI the order is complete
		err = SendResponseToUI(ctx, req, CompleteStage)
		if err != nil {
			return err
		}

		break
	}

	return nil
}

func WorkflowNameForEmail(email string) string {
	return "ui-workflow-" + email
}

func RegisterEmail(ctx context.Context, email string) error {
	logger := activity.GetLogger(ctx)

	logger.Info("activity: registering email", email)

	return nil
}

func ValidateSize(ctx context.Context, size string) error {
	if size != "small" && size != "medium" && size != "large" {
		return fmt.Errorf("size: %s is not valid", size)
	}

	return nil
}

func ValidateColor(ctx context.Context, color string) error {
	if color != "red" && color != "blue" {
		return fmt.Errorf("color: %s is not valid", color)
	}

	return nil
}

func StartOrderWorkflow(ctx workflow.Context, email string) (OrderStatus, error) {
	orderWorkflowID := WorkflowNameForEmail(email)
	status := OrderStatus{OrderID: orderWorkflowID}

	cwo := workflow.ChildWorkflowOptions{
		// Here we force a consistent workflow ID for a given email address
		// This prevents multiple concurrent orders for the same email
		WorkflowID:               orderWorkflowID,
		ParentClosePolicy:        enumspb.PARENT_CLOSE_POLICY_ABANDON,
		WorkflowExecutionTimeout: time.Minute * 30,
	}
	ctx = workflow.WithChildOptions(ctx, cwo)

	orderWorkflow := workflow.ExecuteChildWorkflow(ctx, OrderWorkflow, email)
	err := orderWorkflow.GetChildWorkflowExecution().Get(ctx, nil)
	if err != nil && !temporal.IsWorkflowExecutionAlreadyStartedError(err) {
		return status, err
	}

	err = SendRequestToOrderWorkflow(ctx, orderWorkflowID, RegisterStage, email)
	if err != nil {
		return status, err
	}

	res, err := ReceiveResponseFromOrderWorkflow(ctx)
	if err != nil {
		return status, err
	}

	status.Stage = res.Stage

	return status, err
}

func RecordSizeWorkflow(ctx workflow.Context, orderWorkflowID string, size string) (OrderStatus, error) {
	status := OrderStatus{OrderID: orderWorkflowID, Stage: SizeStage}

	err := SendRequestToOrderWorkflow(ctx, orderWorkflowID, SizeStage, size)
	if err != nil {
		return status, err
	}

	res, err := ReceiveResponseFromOrderWorkflow(ctx)
	if err != nil {
		return status, res.Error
	}

	status.Stage = res.Stage

	return status, nil
}

func RecordColorWorkflow(ctx workflow.Context, orderWorkflowID string, color string) (OrderStatus, error) {
	status := OrderStatus{OrderID: orderWorkflowID, Stage: ColorStage}

	err := SendRequestToOrderWorkflow(ctx, orderWorkflowID, ColorStage, color)
	if err != nil {
		return status, err
	}

	res, err := ReceiveResponseFromOrderWorkflow(ctx)
	if err != nil {
		return status, res.Error
	}

	status.Stage = res.Stage

	return status, nil
}
