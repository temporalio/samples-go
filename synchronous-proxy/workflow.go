package synchronousproxy

import (
	"time"

	"github.com/temporalio/samples-go/synchronous-proxy/proxy"
	"go.temporal.io/sdk/workflow"
)

const (
	RegisterStage = "register"
	SizeStage     = "size"
	ColorStage    = "color"
	ShippingStage = "shipping"
)

var (
	TShirtSizes = []string{
		"small",
		"medium",
		"large",
	}

	TShirtColors = []string{
		"red",
		"blue",
		"black",
	}
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
func OrderWorkflow(ctx workflow.Context) error {
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

		// Tell the UI the order is pending shipping
		err = proxy.SendResponse(ctx, id, ShippingStage, "")
		if err != nil {
			return err
		}

		break
	}

	cw := workflow.ExecuteChildWorkflow(ctx, ShippingWorkflow, order)
	err := cw.Get(ctx, nil)
	if err != nil {
		return err
	}

	return nil
}

func ShippingWorkflow(ctx workflow.Context, order TShirtOrder) error {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var deliveryDate time.Time

	err := workflow.ExecuteActivity(ctx, ScheduleDelivery, order).Get(ctx, &deliveryDate)
	if err != nil {
		return err
	}

	err = workflow.ExecuteActivity(ctx, SendDeliveryEmail, order, deliveryDate).Get(ctx, nil)
	if err != nil {
		return err
	}

	return nil
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
