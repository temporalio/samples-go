package shoppingcart

import (
	"fmt"
	"go.temporal.io/sdk/workflow"
)

var (
	UpdateName    = "shopping-cart"
	TaskQueueName = "shopping-cart-tq"
)

type CartState struct {
	Items map[string]int // itemID -> quantity
}

func CartWorkflow(ctx workflow.Context, cart *CartState) error {
	if cart == nil {
		cart = &CartState{make(map[string]int)}
	}
	logger := workflow.GetLogger(ctx)

	if err := workflow.SetUpdateHandlerWithOptions(ctx, UpdateName, func(ctx workflow.Context, actionType string, itemID string) (*CartState, error) {
		logger.Info("Received update,", actionType, itemID)
		switch actionType {
		case "add":
			cart.Items[itemID] += 1
		case "remove":
			cart.Items[itemID] -= 1
			if cart.Items[itemID] <= 0 {
				delete(cart.Items, itemID)
			}
		case "list":
		default:
			logger.Error("Unsupported action type.")
		}

		return cart, nil
	}, workflow.UpdateHandlerOptions{
		Validator: func(ctx workflow.Context, actionType string, itemId string) error {
			switch actionType {
			case "add", "remove":
				if itemId == "" {
					return fmt.Errorf("itemId must be specified for add or remove actionType")
				}
			case "list":
				if itemId != "" {
					logger.Warn("ItemId not needed for \"list\" actionType.")
				}
			default:
				return fmt.Errorf("unsupported action type: %s", actionType)
			}
			return nil
		},
	}); err != nil {
		return err
	}

	signalChan := workflow.GetSignalChannel(ctx, "checkout")

	err := workflow.Await(ctx, func() bool { return workflow.GetInfo(ctx).GetContinueAsNewSuggested() || signalChan.Receive(ctx, nil) })
	if err != nil {
		return err
	}
	if workflow.GetInfo(ctx).GetContinueAsNewSuggested() {
		err := workflow.Await(ctx, func() bool {
			return workflow.AllHandlersFinished(ctx)
		})
		if err != nil {
			return err
		}
		logger.Info("Continuing as new")

		return workflow.NewContinueAsNewError(ctx, CartWorkflow, cart)
	}

	return nil

}
