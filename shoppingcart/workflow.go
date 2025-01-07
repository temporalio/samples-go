package shoppingcart

import (
	"fmt"
	"go.temporal.io/sdk/workflow"
)

var (
	shoppingServerHostPort = "http://localhost:8099"
)

type CartSignalPayload struct {
	Action   string `json:"action"` // "add" or "remove"
	ItemID   string `json:"item_id"`
	Quantity int    `json:"quantity"`
}

type CartState map[string]int // itemID -> quantity

func CartWorkflow(ctx workflow.Context) error {
	cart := make(CartState)
	cart["apple"] = 1

	// Signal channel for cart updates
	signalChannel := workflow.GetSignalChannel(ctx, "cart_signal")

	// Register a query handler to get the cart state
	workflow.SetQueryHandler(ctx, "get_cart", func() (CartState, error) {
		return cart, nil
	})

	workflow.GetLogger(ctx).Info("CartWorkflow started. Listening for signals...")

	// Listen for signals and update the cart state in a loop
	for {
		var payload CartSignalPayload
		fmt.Println("[SignalPayload]", payload)
		// Block until a signal is received
		signalChannel.Receive(ctx, &payload)

		// Process the received signal
		switch payload.Action {
		case "add":
			if payload.Quantity <= 0 {
				delete(cart, payload.ItemID)
			} else {
				cart[payload.ItemID] += payload.Quantity
			}
			workflow.GetLogger(ctx).Info("Item added to cart", "item_id", payload.ItemID, "quantity", payload.Quantity)

		case "remove":
			delete(cart, payload.ItemID)
			workflow.GetLogger(ctx).Info("Item removed from cart", "item_id", payload.ItemID)

		default:
			workflow.GetLogger(ctx).Warn("Unknown action received", "action", payload.Action)
		}

		// Yield control to allow Temporal to process other tasks
		//workflow.Yield(ctx)
	}

	// This return statement is unreachable because the loop runs indefinitely.
	// You can add logic to break the loop if needed (e.g., based on a "stop" signal).
	return nil
}
