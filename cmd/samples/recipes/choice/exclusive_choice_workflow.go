package main

import (
	"fmt"
	"math/rand"
	"time"

	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/workflow"
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
	workflow.Register(ExclusiveChoiceWorkflow)
	activity.Register(getOrderActivity)
	activity.Register(orderAppleActivity)
	activity.Register(orderBananaActivity)
	activity.Register(orderCherryActivity)
	activity.Register(orderOrangeActivity)
}

// ExclusiveChoiceWorkflow Workflow Decider.
func ExclusiveChoiceWorkflow(ctx workflow.Context) error {
	// Get order.
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
		HeartbeatTimeout:       time.Second * 20,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var orderChoice string
	err := workflow.ExecuteActivity(ctx, getOrderActivity).Get(ctx, &orderChoice)
	if err != nil {
		return err
	}

	logger := workflow.GetLogger(ctx)

	// choose next activity based on order result
	switch orderChoice {
	case orderChoiceApple:
		workflow.ExecuteActivity(ctx, orderAppleActivity, orderChoice)
	case orderChoiceBanana:
		workflow.ExecuteActivity(ctx, orderBananaActivity, orderChoice)
	case orderChoiceCherry:
		workflow.ExecuteActivity(ctx, orderCherryActivity, orderChoice)
	case orderChoiceOrange:
		workflow.ExecuteActivity(ctx, orderOrangeActivity, orderChoice)
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
