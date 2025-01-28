package shoppingcart

import (
	"fmt"
	"go.temporal.io/sdk/workflow"
)

var (
	UpdateName    = "shopping-cart"
	TaskQueueName = "shopping-cart-tq"
)

type CartSignalPayload struct {
	Action   string `json:"action"` // "add" or "remove"
	ItemID   string `json:"item_id"`
	Quantity int    `json:"quantity"`
}

type CartState struct {
	Items map[string]int // itemID -> quantity
}

func CartWorkflow(ctx workflow.Context) error {
	cart := CartState{make(map[string]int)}
	logger := workflow.GetLogger(ctx)
	var checkout bool

	if err := workflow.SetUpdateHandlerWithOptions(ctx, UpdateName, func(ctx workflow.Context, actionType string, itemID string) (CartState, error) {
		logger.Info("Received update,", actionType, itemID)
		if actionType == "checkout" {
			checkout = true
		}
		if itemID != "" {
			if actionType == "add" {
				cart.Items[itemID] += 1
			} else if actionType == "remove" {
				cart.Items[itemID] -= 1
				if cart.Items[itemID] <= 0 {
					delete(cart.Items, itemID)
				}
			}
		}
		return cart, nil
	}, workflow.UpdateHandlerOptions{
		Validator: func(ctx workflow.Context, actionType string, itemID string) error {
			switch actionType {
			case "add", "remove", "checkout", "":
				return nil
			default:
				return fmt.Errorf("unsupported action type: %s", actionType)
			}
		},
	}); err != nil {
		return err
	}

	err := workflow.Await(ctx, func() bool { return workflow.GetInfo(ctx).GetContinueAsNewSuggested() || checkout })
	if err != nil {
		return err
	}
	if workflow.GetInfo(ctx).GetContinueAsNewSuggested() {
		logger.Info("Continuing as new")
		return workflow.NewContinueAsNewError(ctx, CartWorkflow)
	}
	if checkout {
		return nil
	}

	return fmt.Errorf("unreachable")

}
