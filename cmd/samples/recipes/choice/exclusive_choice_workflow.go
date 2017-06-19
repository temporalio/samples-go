package main

import (
	"fmt"
	"math/rand"
	"time"

	"go.uber.org/cadence"
	"go.uber.org/zap"
)

/**
 * This sample workflow Execute one of many code paths based on the result of an activity.
 */

const (
	// ApplicationName is the task list for this sample
	ApplicationName = "choiceGroup"

	orderChoiceApple  = "apple"
	orderChoiceBanana = "banana"
	orderChoiceCherry = "cherry"
	orderChoiceOrange = "orange"
)

var _orderChoices = []string{orderChoiceApple, orderChoiceBanana, orderChoiceCherry, orderChoiceOrange}

// This is registration process where you register all your workflows
// and activity function handlers.
func init() {
	cadence.RegisterWorkflow(ExclusiveChoiceWorkflow)
	cadence.RegisterActivity(getOrderActivity)
	cadence.RegisterActivity(orderAppleActivity)
	cadence.RegisterActivity(orderBananaActivity)
	cadence.RegisterActivity(orderCherryActivity)
	cadence.RegisterActivity(orderOrangeActivity)
}

// ExclusiveChoiceWorkflow Workflow Decider.
func ExclusiveChoiceWorkflow(ctx cadence.Context) error {
	// Get order.
	ao := cadence.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
		HeartbeatTimeout:       time.Second * 20,
	}
	ctx = cadence.WithActivityOptions(ctx, ao)

	var orderChoice string
	err := cadence.ExecuteActivity(ctx, getOrderActivity).Get(ctx, &orderChoice)
	if err != nil {
		return err
	}

	logger := cadence.GetLogger(ctx)

	// choose next activity based on order result
	switch orderChoice {
	case orderChoiceApple:
		cadence.ExecuteActivity(ctx, orderAppleActivity, orderChoice)
	case orderChoiceBanana:
		cadence.ExecuteActivity(ctx, orderBananaActivity, orderChoice)
	case orderChoiceCherry:
		cadence.ExecuteActivity(ctx, orderCherryActivity, orderChoice)
	case orderChoiceOrange:
		cadence.ExecuteActivity(ctx, orderOrangeActivity, orderChoice)
	default:
		logger.Error("Unexpected order", zap.String("Choice", orderChoice))
	}

	logger.Info("Workflow completed.")
	return nil
}

func getOrderActivity() (string, error) {
	idx := rand.Intn(len(_orderChoices))
	order := _orderChoices[idx]
	fmt.Printf("Order is for %s\n", order)
	return order, nil
}

func orderAppleActivity(choice string) error {
	fmt.Printf("Order choice: %v\n", choice)
	return nil
}

func orderBananaActivity(choice string) error {
	fmt.Printf("Order choice: %v\n", choice)
	return nil
}

func orderCherryActivity(choice string) error {
	fmt.Printf("Order choice: %v\n", choice)
	return nil
}

func orderOrangeActivity(choice string) error {
	fmt.Printf("Order choice: %v\n", choice)
	return nil
}
