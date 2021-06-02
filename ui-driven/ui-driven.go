package uidriven

import (
	"time"

	"github.com/temporalio/samples-go/ui-driven/proxy"
	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

const (
	RegisterStage = "register"
	SizeStage     = "size"
	ColorStage    = "color"
	CompleteStage = "complete"
)

type TShirtOrder struct {
	Email string
	Size  string
	Color string
}

type OrderStatus struct {
	OrderID string
	Stage   string
}

// Workflow is a workflow driven by interaction from a UI.
func OrderWorkflow(ctx workflow.Context, email string) error {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	order := TShirtOrder{}

	// Loop until we receive a valid email
	for {
		id, _, email := proxy.ReceiveRequest(ctx)

		err := workflow.ExecuteActivity(ctx, RegisterEmail, email).Get(ctx, nil)
		if err != nil {
			err := proxy.SendErrorResponse(ctx, id, err)
			if err != nil {
				return err
			}

			continue
		}

		order.Email = email

		err = proxy.SendResponse(ctx, id, SizeStage, "")
		if err != nil {
			return err
		}

		break
	}

	// Loop until we receive a valid size
	for {
		id, _, size := proxy.ReceiveRequest(ctx)

		err := workflow.ExecuteActivity(ctx, ValidateSize, size).Get(ctx, nil)
		if err != nil {
			err := proxy.SendErrorResponse(ctx, id, err)
			if err != nil {
				return err
			}

			continue
		}

		order.Size = size

		err = proxy.SendResponse(ctx, id, ColorStage, "")
		if err != nil {
			return err
		}

		break
	}

	// Loop until we receive a valid color
	for {
		id, _, color := proxy.ReceiveRequest(ctx)

		err := workflow.ExecuteActivity(ctx, ValidateColor, color).Get(ctx, nil)
		if err != nil {
			err := proxy.SendErrorResponse(ctx, id, err)
			if err != nil {
				return err
			}

			continue
		}

		order.Color = color

		// Tell the UI the order is complete
		err = proxy.SendResponse(ctx, id, CompleteStage, "")
		if err != nil {
			return err
		}

		break
	}

	err := workflow.ExecuteActivity(ctx, ProcessOrder, order).Get(ctx, nil)
	if err != nil {
		return err
	}

	return nil
}

func WorkflowNameForEmail(email string) string {
	return "ui-workflow-" + email
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

	err = proxy.SendRequest(ctx, orderWorkflowID, RegisterStage, email)
	if err != nil {
		return status, err
	}

	stage, _, err := proxy.ReceiveResponse(ctx)
	if err != nil {
		return status, err
	}

	status.Stage = stage

	return status, err
}

func UpdateOrderWorkflow(ctx workflow.Context, orderWorkflowID string, stage string, value string) (OrderStatus, error) {
	status := OrderStatus{OrderID: orderWorkflowID, Stage: stage}

	err := proxy.SendRequest(ctx, orderWorkflowID, stage, value)
	if err != nil {
		return status, err
	}

	nextStage, _, err := proxy.ReceiveResponse(ctx)
	if err != nil {
		return status, err
	}

	status.Stage = nextStage

	return status, nil
}
