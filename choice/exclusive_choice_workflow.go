package choice

import (
	"fmt"
	"math/rand"
	"time"

	"go.temporal.io/temporal/workflow"
	"go.uber.org/zap"
)

/**
 * This sample workflow Execute one of many code paths based on the result of an activity.
 */

const (
	orderChoiceApple  = "apple"
	orderChoiceBanana = "banana"
	orderChoiceCherry = "cherry"
	orderChoiceOrange = "orange"
)

var _orderChoices = []string{orderChoiceApple, orderChoiceBanana, orderChoiceCherry, orderChoiceOrange}

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
	err := workflow.ExecuteActivity(ctx, GetOrderActivity).Get(ctx, &orderChoice)
	if err != nil {
		return err
	}

	logger := workflow.GetLogger(ctx)

	// choose next activity based on order result
	switch orderChoice {
	case orderChoiceApple:
		workflow.ExecuteActivity(ctx, OrderAppleActivity, orderChoice)
	case orderChoiceBanana:
		workflow.ExecuteActivity(ctx, OrderBananaActivity, orderChoice)
	case orderChoiceCherry:
		workflow.ExecuteActivity(ctx, OrderCherryActivity, orderChoice)
	case orderChoiceOrange:
		workflow.ExecuteActivity(ctx, OrderOrangeActivity, orderChoice)
	default:
		logger.Error("Unexpected order", zap.String("Choice", orderChoice))
	}

	logger.Info("Workflow completed.")
	return nil
}

func GetOrderActivity() (string, error) {
	idx := rand.Intn(len(_orderChoices))
	order := _orderChoices[idx]
	fmt.Printf("Order is for %s\n", order)
	return order, nil
}

func OrderAppleActivity(choice string) error {
	fmt.Printf("Order choice: %v\n", choice)
	return nil
}

func OrderBananaActivity(choice string) error {
	fmt.Printf("Order choice: %v\n", choice)
	return nil
}

func OrderCherryActivity(choice string) error {
	fmt.Printf("Order choice: %v\n", choice)
	return nil
}

func OrderOrangeActivity(choice string) error {
	fmt.Printf("Order choice: %v\n", choice)
	return nil
}
