package shoppingcart

import (
	"fmt"
	"go.temporal.io/sdk/workflow"
	"log"
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

type CartState map[string]int // itemID -> quantity

func CartWorkflow(ctx workflow.Context) error {
	cart := make(CartState)

	if err := workflow.SetUpdateHandler(ctx, UpdateName, func(ctx workflow.Context, actionType string, itemID string) (CartState, error) {
		fmt.Println("Received update,", actionType, itemID)
		if itemID != "" {
			if actionType == "add" {
				cart[itemID] += 1
			} else if actionType == "remove" {
				cart[itemID] -= 1
				if cart[itemID] <= 0 {
					delete(cart, itemID)
				}
			} else {
				log.Fatalln("Unknown action type:", actionType)
			}
		}
		return cart, nil
	}); err != nil {
		return err
	}

	// Keep workflow alive to continue to listen receive updates.
	return workflow.Await(ctx, func() bool { return false })
}
